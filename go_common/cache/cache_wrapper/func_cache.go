package cache_wrapper

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/luulethe/quiz/go_common/cache"
	cache_encoder "github.com/luulethe/quiz/go_common/cache/encoder"
	"github.com/luulethe/quiz/go_common/log"
)

type Encoder interface {
	Encode(interface{}) (string, error)
	Decode(string, interface{}) error
}

var defaultEncoder = cache_encoder.NewJSONEncoder()

type WrapperConfig struct {
	KeyFormat  string
	Expire     time.Duration
	DataType   interface{}
	CacheType  string
	Encoder    Encoder
	SliceInput bool
	Compress   bool
}

func (c *WrapperConfig) GetEncoder() Encoder {
	encoder := c.Encoder
	if encoder == nil {
		encoder = defaultEncoder
	}
	return encoder
}

var cacheInstanceMap = map[string]cache.SimpleCache{}

func RegisterCacheType(name string, cacheInstance cache.SimpleCache) {
	cacheInstanceMap[name] = cacheInstance
}

func getFromCache(ctx context.Context, config *WrapperConfig, key interface{}, resultType reflect.Type) (reflect.Value, error) {
	cacheInstance := cacheInstanceMap[config.CacheType]
	cacheKey := fmt.Sprintf(config.KeyFormat, key)

	var result reflect.Value
	resultString, err := cacheInstance.Get(cacheKey)
	if err != nil {
		return result, err
	}
	if config.Compress {
		resultString, err = decompressString(resultString.(string))
		if err != nil {
			log.Errorf(ctx, "decompress_fail|cacheKey=%s", cacheKey)
			return result, err
		}
	}

	if config.Compress || cacheInstance.NeedEncode() {
		result = reflect.New(resultType)
		err = config.GetEncoder().Decode(resultString.(string), result.Interface())
	} else {
		result = reflect.ValueOf(resultString)
	}
	if err != nil {
		log.Errorf(ctx, "decode_fail|cacheKey=%s", cacheKey)
	}
	return result, err
}

func mGetFromCache(
	ctx context.Context, config *WrapperConfig, keys []interface{}, resultType reflect.Type,
) (map[interface{}]reflect.Value, error) {
	cacheInstance := cacheInstanceMap[config.CacheType]

	cacheKeys := []string{}
	for _, key := range keys {
		cacheKeys = append(cacheKeys, fmt.Sprintf(config.KeyFormat, key))
	}

	results, err := cacheInstance.MGet(cacheKeys...)
	if err != nil {
		return nil, err
	}

	resultMap := map[interface{}]reflect.Value{}
	for i, r := range results {
		if r == nil || i >= len(keys) {
			continue
		}
		key := keys[i]

		resultString := r
		if config.Compress {
			resultString, err = decompressString(resultString.(string))
			if err != nil {
				log.Errorf(ctx, "decompress_fail|cacheKey=%s", cacheKeys[i])
				continue
			}
		}

		if config.Compress || cacheInstance.NeedEncode() {
			result := reflect.New(resultType)
			err = config.GetEncoder().Decode(resultString.(string), result.Interface())
			if err != nil {
				log.Errorf(ctx, "decode_fail|cacheKey=%s", cacheKeys[i])
				continue
			}
			resultMap[key] = result
		} else {
			resultMap[key] = reflect.ValueOf(resultString)
		}
	}
	return resultMap, err
}

func setToCache(ctx context.Context, config *WrapperConfig, key interface{}, result interface{}) (err error) {
	cacheInstance := cacheInstanceMap[config.CacheType]
	cacheKey := fmt.Sprintf(config.KeyFormat, key)

	resultString := result
	if config.Compress || cacheInstance.NeedEncode() {
		resultString, err = config.GetEncoder().Encode(result)
		if err != nil {
			log.Errorf(ctx, "encode_fail|cacheKey=%s", cacheKey)
			return err
		}
	}
	if config.Compress {
		resultString, err = compressString(resultString.(string))
		if err != nil {
			log.Errorf(ctx, "compress_fail|cacheKey=%s", cacheKey)
			return err
		}
	}
	return cacheInstance.Set(cacheKey, resultString, config.Expire)
}

func mSetToCache(ctx context.Context, config *WrapperConfig, pairs map[interface{}]interface{}) (err error) {
	cacheInstance := cacheInstanceMap[config.CacheType]

	cachePairs := map[string]interface{}{}
	for key, result := range pairs {
		cacheKey := fmt.Sprintf(config.KeyFormat, key)

		resultString := result
		if config.Compress || cacheInstance.NeedEncode() {
			resultString, err = config.GetEncoder().Encode(result)
			if err != nil {
				log.Errorf(ctx, "encode_fail|cacheKey=%s", cacheKey)
				return err
			}
		}
		if config.Compress {
			resultString, err = compressString(resultString.(string))
			if err != nil {
				log.Errorf(ctx, "compress_fail|cacheKey=%s", cacheKey)
				return err
			}
		}
		cachePairs[cacheKey] = resultString
	}
	return cacheInstance.MSet(cachePairs, config.Expire)
}

func WithCache(config *WrapperConfig, input interface{}, f func(interface{}) (interface{}, error)) (interface{}, error) {
	if config.SliceInput {
		return WithMultiKey(config, input, f)
	}
	return WithSingleKey(config, input, f)
}

func WithSingleKey(config *WrapperConfig, input interface{}, f func(interface{}) (interface{}, error)) (interface{}, error) {
	ctx := context.Background()
	resultType := reflect.TypeOf(config.DataType).Elem()

	result, err := getFromCache(ctx, config, input, resultType)
	if err == nil {
		return result.Interface(), nil
	}
	funcResult, err := f(input)
	if err == nil {
		_ = setToCache(ctx, config, input, funcResult)
	}
	return funcResult, err
}

func WithMultiKey(config *WrapperConfig, inputList interface{}, f func(interface{}) (interface{}, error)) (interface{}, error) {
	ctx := context.Background()
	resultType := reflect.TypeOf(config.DataType).Elem()
	inputListValue := reflect.ValueOf(inputList)
	if inputListValue.Kind() != reflect.Slice {
		log.Errorf(ctx, "input_is_not_list|inputValue.Kind()=%s", inputListValue.Kind().String())
		return nil, fmt.Errorf("input_is_not_list")
	}

	var aMapType = reflect.MapOf(inputListValue.Type().Elem(), reflect.TypeOf(config.DataType))
	resultMap := reflect.MakeMapWithSize(aMapType, inputListValue.Len())
	if inputListValue.Len() == 0 {
		return resultMap.Interface(), nil
	}

	var keys []interface{}
	missingKeys := reflect.MakeSlice(inputListValue.Type(), 0, 0)
	for i := 0; i < inputListValue.Len(); i++ {
		keys = append(keys, inputListValue.Index(i).Interface())
	}
	cacheResult, err := mGetFromCache(ctx, config, keys, resultType)
	if err != nil {
		log.Errorf(ctx, "mGetFromCache_fail|err:%s", err.Error())
		missingKeys = inputListValue
	} else {
		for i := 0; i < inputListValue.Len(); i++ {
			inputValue := inputListValue.Index(i)
			if value, ok := cacheResult[inputValue.Interface()]; ok {
				resultMap.SetMapIndex(inputValue, value)
			} else {
				missingKeys = reflect.Append(missingKeys, inputValue)
			}
		}
	}

	if missingKeys.Len() > 0 {
		missingResults, err := f(missingKeys.Interface())
		if err != nil {
			return nil, err
		}

		v := reflect.ValueOf(missingResults)
		if v.Kind() == reflect.Map {
			pairs := map[interface{}]interface{}{}
			for _, key := range v.MapKeys() {
				value := v.MapIndex(key)
				resultMap.SetMapIndex(key, value)
				pairs[key.Interface()] = value.Interface()
			}
			err := mSetToCache(ctx, config, pairs)
			if err != nil {
				log.Errorf(ctx, "mSetToCache_fail|err=%s", err.Error())
			}
		} else {
			log.Errorf(ctx, "result_is_not_map|v.Kind()=%s", v.Kind().String())
		}
	}

	return resultMap.Interface(), nil
}
