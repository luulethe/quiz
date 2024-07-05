package cache

import (
	"time"

	"github.com/go-redis/redis"
)

type Reader interface {
	// Simple K/Vs
	Get(key string) (interface{}, error)
	MGet(keys ...string) ([]interface{}, error)
	Exists(key string) (int64, error)
	Keys(key string) ([]string, error)

	// ScanAll has almost same functionality as Keys.
	// Internally it use scan with per scan count = 100, so slower but without blocking redis
	ScanAll(key string) ([]string, error)

	// HashMap
	HGet(key, field string) (interface{}, error)
	HExists(key, field string) (bool, error)
	HGetAll(key string) (map[string]string, error)

	// Set
	SMembers(key string) ([]string, error)
	SIsMember(key, member string) (bool, error)

	// Queue/List
	LLen(key string) (int64, error)
	LIndex(key string, index int64) (string, error)
	LRange(key string, from int64, to int64) ([]string, error)
}

type Writer interface {
	// Simple K/Vs
	// Set will set key to value, zero expiration means the key has no expiration time.
	Set(key, value string, expire time.Duration) (err error)
	MSet(pairs map[string]string, expiration time.Duration) (err error)
	Incr(key string) (value int64, err error)
	SetNX(key string, value interface{}, expiration time.Duration) (bool, error)
	GetSet(key string, value interface{}) (string, error)
	Del(key ...string) (count int64, err error)
	Expire(key string, expiration time.Duration) (bool, error)
	ExpireAt(key string, tm time.Time) (bool, error)

	// HashMap
	HDel(key, field string) (count int64, err error)
	HSet(key, field, value string) (bool, error)
	HIncr(key, field string) (value int64, err error)
	HMSet(key string, fields map[string]interface{}) (string, error)

	// Set
	SAdd(key string, members ...interface{}) (int64, error)
	SRem(key string, members ...interface{}) (int64, error)

	// Queue
	LPush(key string, values ...interface{}) error
	RPush(key string, values ...interface{}) error
	LPop(key string) (string, error)
	RPop(key string) (string, error)
	RPopLPush(src, dest string) (string, error)
	BLPop(timeout time.Duration, key string) (string, error)
	BRPop(timeout time.Duration, key string) (string, error)
	BRPopLPush(src, dest string, timeout time.Duration) (string, error)
	LRem(key string, count int64, val string) (int64, error)

	// Eval
	Eval(script string, key, args []string) (string, error)
}

type EnhancedCache interface {
	Reader
	Writer

	SetBit(key string, pos int64, value int) (res int64, err error)
	GetBit(key string, pos int64) (value int64, err error)
	BitCount(key string) (value int64, err error)

	// Break the encapsulation, but I think the raw api is good enough, and worth it.
	Client() *redis.Client
}

type SimpleCache interface {
	NeedEncode() bool
	Get(key string) (interface{}, error)
	MGet(keys ...string) ([]interface{}, error)
	Set(key string, value interface{}, expire time.Duration) (err error)
	MSet(pairs map[string]interface{}, expiration time.Duration) (err error)
}
