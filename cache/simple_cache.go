package cache

import (
	"sync"
	"time"
)

// Temporary data stored in cache with expiration.
type tempData struct {
	expiration time.Time
	value      interface{}
}

// In memory cache storage.
type SimpleCache struct {
	sync.RWMutex
	cache map[string]tempData
}

func New() *SimpleCache {
	return &SimpleCache{cache: make(map[string]tempData)}
}

// Add to cache.
// TODO: Limit for the number of elements or the size fo the storage.
func (c *SimpleCache) Set(key string, value interface{}, ttl time.Duration) {
	c.Lock()
	defer c.Unlock()

	c.cache[key] = tempData{
		expiration: time.Now().Add(ttl),
		value:      value,
	}
}

// Get value from cache, with extra boolean showing cache hit or miss.
func (c *SimpleCache) Get(key string) (interface{}, bool) {
	c.Lock()
	defer c.Unlock()

	data, ok := c.cache[key]
	if !ok {
		return "", false
	}

	if time.Now().After(data.expiration) {
		delete(c.cache, key)
		return "", false
	}

	return data.value, true
}
