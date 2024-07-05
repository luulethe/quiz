package cache

import (
	"errors"
	"time"

	"github.com/go-redis/redis"
)

type RedisCache struct {
	client *redis.Client
}

type RedisOption struct {
	Password     string
	DB           int
	PoolSize     int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func DefaultRedisOption(db int, poolSize int, timeout time.Duration) *RedisOption {
	return &RedisOption{
		DB:           db,
		PoolSize:     poolSize,
		DialTimeout:  timeout,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}
}

func NewRedisClient(address string, option *RedisOption) (*RedisCache, error) {
	client := redis.NewClient(
		&redis.Options{
			Network:      "tcp4",
			Addr:         address,
			Password:     option.Password,
			DB:           option.DB,
			PoolSize:     option.PoolSize,
			DialTimeout:  option.DialTimeout,
			ReadTimeout:  option.ReadTimeout,
			WriteTimeout: option.WriteTimeout,
			IdleTimeout:  option.IdleTimeout,
		})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}
	return &RedisCache{
		client: client,
	}, nil
}

func NewFailoverRedisClient(masterName string, sentinelAddresses []string, option *RedisOption) (*RedisCache, error) {
	client := redis.NewFailoverClient(
		&redis.FailoverOptions{
			MasterName:    masterName,
			SentinelAddrs: sentinelAddresses,
			Password:      option.Password,
			DB:            option.DB,
			PoolSize:      option.PoolSize,
			DialTimeout:   option.DialTimeout,
			ReadTimeout:   option.ReadTimeout,
			WriteTimeout:  option.WriteTimeout,
			IdleTimeout:   option.IdleTimeout,
		})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}
	return &RedisCache{
		client: client,
	}, nil
}

func (c RedisCache) NeedEncode() bool {
	return true
}

func (c RedisCache) PoolStats() *redis.PoolStats {
	return c.client.PoolStats()
}

func (c RedisCache) Get(key string) (interface{}, error) {
	return c.client.Get(key).Result()
}

func (c RedisCache) Set(key string, value interface{}, expire time.Duration) (err error) {
	status := c.client.Set(key, value, expire)
	return status.Err()
}

func (c RedisCache) SetNX(key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.client.SetNX(key, value, expiration).Result()
}

func (c RedisCache) GetSet(key string, value interface{}) (string, error) {
	return c.client.GetSet(key, value).Result()
}

func (c RedisCache) Del(keys ...string) (count int64, err error) {
	return c.client.Del(keys...).Result()
}

func (c RedisCache) Exists(key string) (int64, error) {
	return c.client.Exists(key).Result()
}

func (c RedisCache) MGet(keys ...string) (values []interface{}, err error) {
	return c.client.MGet(keys...).Result()
}

func (c RedisCache) MSet(pairs map[string]interface{}, expiration time.Duration) (err error) {
	pipeline := c.client.TxPipeline()
	for key, value := range pairs {
		pipeline.Set(key, value, expiration)
	}
	_, err = pipeline.Exec()
	return
}

func (c RedisCache) Close() error {
	return c.client.Close()
}

func (c RedisCache) HGet(key, field string) (interface{}, error) {
	return c.client.HGet(key, field).Result()
}

func (c RedisCache) HGetAll(key string) (map[string]string, error) {
	return c.client.HGetAll(key).Result()
}

func (c RedisCache) HExists(key, field string) (bool, error) {
	return c.client.HExists(key, field).Result()
}

func (c RedisCache) HDel(key, field string) (int64, error) {
	return c.client.HDel(key, field).Result()
}

func (c RedisCache) HSet(key, field, value string) (bool, error) {
	return c.client.HSet(key, field, value).Result()
}

func (c RedisCache) HMSet(key string, fields map[string]interface{}) (string, error) {
	if len(fields) < 1 {
		return "", errors.New("Invalid Argument")
	}
	return c.client.HMSet(key, fields).Result()
}

func (c RedisCache) Incr(key string) (value int64, err error) {
	return c.client.Incr(key).Result()
}

func (c RedisCache) HIncr(key, field string) (value int64, err error) {
	return c.client.HIncrBy(key, field, 1).Result()
}

func (c RedisCache) SetBit(key string, pos int64, value int) (res int64, err error) {
	return c.client.SetBit(key, pos, value).Result()
}

func (c RedisCache) GetBit(key string, pos int64) (value int64, err error) {
	return c.client.GetBit(key, pos).Result()
}

func (c RedisCache) BitCount(key string) (value int64, err error) {
	return c.client.BitCount(key, nil).Result()
}

func (c RedisCache) SMembers(key string) ([]string, error) {
	return c.client.SMembers(key).Result()
}

func (c RedisCache) SAdd(key string, members ...interface{}) (int64, error) {
	return c.client.SAdd(key, members...).Result()
}

func (c RedisCache) SRem(key string, members ...interface{}) (int64, error) {
	return c.client.SRem(key, members...).Result()
}

func (c RedisCache) SIsMember(key, member string) (bool, error) {
	return c.client.SIsMember(key, member).Result()
}

func (c RedisCache) RPush(key string, values ...interface{}) error {
	return c.client.RPush(key, values...).Err()
}

func (c RedisCache) LPush(key string, values ...interface{}) error {
	return c.client.LPush(key, values...).Err()
}

func (c RedisCache) LPop(key string) (string, error) {
	return c.client.LPop(key).Result()
}

func (c RedisCache) RPop(key string) (string, error) {
	return c.client.RPop(key).Result()
}

func (c RedisCache) RPopLPush(source, destination string) (string, error) {
	return c.client.RPopLPush(source, destination).Result()
}

func (c RedisCache) Keys(key string) ([]string, error) {
	return c.client.Keys(key).Result()
}

func (c RedisCache) ScanAll(key string) ([]string, error) {
	var res []string

	// 100 is a reasonable count, we hard-code this to simplify API.
	const scanCountPerStep = 100

	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = c.client.Scan(cursor, key, scanCountPerStep).Result()
		if err != nil {
			return nil, err
		}
		res = append(res, keys...)
		if cursor == 0 {
			break
		}
	}

	return res, nil
}

func (c RedisCache) Expire(key string, expiration time.Duration) (bool, error) {
	return c.client.Expire(key, expiration).Result()
}

func (c RedisCache) ExpireAt(key string, tm time.Time) (bool, error) {
	return c.client.ExpireAt(key, tm).Result()
}

func (c RedisCache) BLPop(timeout time.Duration, key string) (string, error) {
	msgs, err := c.client.BLPop(timeout, key).Result()
	if err != nil {
		return "", err
	}
	return msgs[1], nil
}

func (c RedisCache) BRPop(timeout time.Duration, key string) (string, error) {
	msgs, err := c.client.BRPop(timeout, key).Result()
	if err != nil {
		return "", err
	}
	return msgs[1], nil
}

func (c RedisCache) BRPopLPush(source, destination string, timeout time.Duration) (string, error) {
	return c.client.BRPopLPush(source, destination, timeout).Result()
}

func (c RedisCache) LRem(key string, count int64, val string) (int64, error) {
	return c.client.LRem(key, count, val).Result()
}

func (c RedisCache) LLen(key string) (int64, error) {
	return c.client.LLen(key).Result()
}

func (c RedisCache) LIndex(key string, index int64) (string, error) {
	return c.client.LIndex(key, index).Result()
}

func (c RedisCache) LRange(key string, start int64, stop int64) ([]string, error) {
	return c.client.LRange(key, start, stop).Result()
}

func (c RedisCache) Eval(script string, keys, args []string) (string, error) {
	val, err := c.client.Eval(script, keys, args).Result()
	if err != nil {
		return "", err
	}
	if result, ok := val.(string); ok {
		return result, nil
	}
	return "", nil
}

func (c RedisCache) Client() *redis.Client {
	return c.client
}
