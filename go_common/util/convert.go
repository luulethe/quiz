package util

import (
	"fmt"
	"strconv"
	"strings"
)

// ListToMapInt32 ...
func ListToMapInt32(list []int32) map[int32]bool {
	m := make(map[int32]bool)
	for _, l := range list {
		m[l] = true
	}
	return m
}

// ListToMapInt64 ...
func ListToMapInt64(list []int64) map[int64]bool {
	m := make(map[int64]bool)
	for _, l := range list {
		m[l] = true
	}
	return m
}

// ListToMapString ...
func ListToMapString(list []string) map[string]bool {
	m := make(map[string]bool)
	for _, l := range list {
		m[l] = true
	}
	return m
}

// MapToListString ...
func MapToListString(m map[string]bool) []string {
	list := make([]string, 0, len(m))
	for t := range m {
		list = append(list, t)
	}
	return list
}

// MapToListInt ...
func MapToListInt(m map[int]bool) []int {
	list := make([]int, 0, len(m))
	for t := range m {
		list = append(list, t)
	}
	return list
}

// MapToListInt32 ...
func MapToListInt32(m map[int32]bool) []int32 {
	list := make([]int32, 0, len(m))
	for t := range m {
		list = append(list, t)
	}
	return list
}

// KVPairFloat32 ... key:string value:float32
type KVPairFloat32 struct {
	Key   string
	Value float32
}

// KVPairFloat32ToMap ...
func KVPairFloat32ToMap(list []KVPairFloat32) map[string]float32 {
	mp := make(map[string]float32, len(list))
	for _, kv := range list {
		mp[kv.Key] = kv.Value
	}
	return mp
}

// StrToFloat32Slice ... convert string slice to float32 slice, if error happens, replace with 0.0
func StrToFloat32Slice(list []string) []float32 {
	data := make([]float32, 0, len(list))
	for _, str := range list {
		num, err := strconv.ParseFloat(str, 32)
		if err != nil {
			num = 0.0
		}
		data = append(data, float32(num))
	}
	return data
}

// Float32SliceToStr ... convert float32 slice to string
func Float32SliceToStr(list []float32, delim string) string {
	data := make([]string, len(list))
	for i, num := range list {
		data[i] = fmt.Sprintf("%f", num)
	}
	return strings.Join(data, delim)
}

// StrToInt32 ...
func StrToInt32(s string) (int32, error) {
	num, err := strconv.ParseInt(s, 10, 32)
	return int32(num), err
}

// StrToInt64 ...
func StrToInt64(s string) (int64, error) {
	num, err := strconv.ParseInt(s, 10, 64)
	return num, err
}

// StrToFloat64 ...
func StrToFloat64(s string) (float64, error) {
	num, err := strconv.ParseFloat(s, 64)
	return num, err
}

// StrToFloat32 ...
func StrToFloat32(s string) (float32, error) {
	num, err := strconv.ParseFloat(s, 32)
	return float32(num), err
}
