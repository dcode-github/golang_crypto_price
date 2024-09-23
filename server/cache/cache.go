package cache

import (
	"cryptoProject/server/model"
	"sync"
	"time"
)

type cacheValue struct {
	value          map[string]float64
	expirationTime time.Time
}

type Cache struct {
	data map[model.CacheKey]cacheValue
	lock sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[model.CacheKey]cacheValue),
	}
}

func (c *Cache) Set(key model.CacheKey, value map[string]float64, expiration time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()

	expirationTime := time.Now().Add(expiration)
	c.data[key] = cacheValue{
		value:          value,
		expirationTime: expirationTime,
	}
}
func (c *Cache) Get(key model.CacheKey) (map[string]float64, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	value, ok := c.data[key]
	if !ok || time.Now().After(value.expirationTime) {
		delete(c.data, key)
		return nil, false
	}
	return value.value, true
}
