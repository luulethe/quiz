package cache

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

type MemoryCache struct {
	*cache.Cache
}

type MemoryCacheOption struct {
	DefaultExpiration time.Duration
	CleanupInterval   time.Duration
}

func NewMemoryCache(option *MemoryCacheOption) (*MemoryCache, error) {
	memoryCache := cache.New(option.DefaultExpiration, option.CleanupInterval)
	return &MemoryCache{memoryCache}, nil
}

func (c *MemoryCache) NeedEncode() bool {
	return false
}

func (c *MemoryCache) Get(key string) (interface{}, error) {
	cacheValue, found := c.Cache.Get(key)
	if !found {
		return nil, fmt.Errorf("memory cache: nil")
	}
	return cacheValue, nil
}

func (c *MemoryCache) Set(key string, value interface{}, expire time.Duration) (err error) {
	c.Cache.Set(key, value, expire)
	return
}

func (c *MemoryCache) MGet(keys ...string) (values []interface{}, err error) {
	for _, key := range keys {
		cacheValue, found := c.Cache.Get(key)
		if found {
			values = append(values, cacheValue)
		} else {
			values = append(values, nil)
		}
	}
	return
}

func (c *MemoryCache) MSet(pairs map[string]interface{}, expiration time.Duration) (err error) {
	for key, value := range pairs {
		err = c.Set(key, value, expiration)
		if err != nil {
			return
		}
	}
	return
}
