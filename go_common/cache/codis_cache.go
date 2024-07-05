package cache

import "time"

type CodisCache struct {
	*RedisCache
}

func NewCodisCache(address string, option *RedisOption) (*CodisCache, error) {
	redisCache, err := NewRedisClient(address, option)
	if err != nil {
		return nil, err
	}
	return &CodisCache{redisCache}, nil
}

func (c CodisCache) MSet(pairs map[string]interface{}, expiration time.Duration) (err error) {
	pipeline := c.client.Pipeline()
	for key, value := range pairs {
		pipeline.Set(key, value, expiration)
	}
	_, err = pipeline.Exec()
	return
}
