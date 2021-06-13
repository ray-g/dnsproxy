package memcache

import (
	"encoding/json"
	"sync"

	"github.com/ray-g/dnsproxy/cache"
	r "github.com/ray-g/dnsproxy/cache/record"
)

// MemoryCache type
type MemoryCache struct {
	sync.RWMutex
	Records  map[string]*r.Record `json:"cache"`
	Capacity int                  `json:"capacity"`
}

func NewCache() cache.Cache {
	return NewSizedCache(0)
}

func NewSizedCache(capacity int) cache.Cache {
	return &MemoryCache{
		Records:  make(map[string]*r.Record),
		Capacity: capacity,
	}
}

func (c *MemoryCache) Set(key string, record *r.Record) error {
	if c.Full() {
		return cache.ErrorCacheFull
	}

	if c.Exists(key) {
		return nil
	}

	c.Lock()
	defer c.Unlock()
	c.Records[key] = record
	return nil
}

func (c *MemoryCache) Get(key string) (record *r.Record, err error) {
	c.RLock()
	record, ok := c.Records[key]
	c.RUnlock()

	if !ok {
		return nil, cache.ErrorCacheKeyMissed
	}

	if record.Expired() {
		c.Remove(key)
		return nil, cache.ErrorCacheKeyExpired
	}

	return record, nil
}

func (c *MemoryCache) Exists(key string) bool {
	c.RLock()
	defer c.RUnlock()

	_, ok := c.Records[key]
	return ok
}

func (c *MemoryCache) Remove(key string) {
	c.Lock()
	defer c.Unlock()

	delete(c.Records, key)
}

func (c *MemoryCache) Length() int {
	c.RLock()
	defer c.RUnlock()

	return len(c.Records)
}

func (c *MemoryCache) Full() bool {
	if c.Capacity == 0 {
		return false
	}
	return c.Length() >= c.Capacity
}

func (c *MemoryCache) Dump() string {
	c.RLock()
	defer c.RUnlock()
	bytes, err := json.Marshal(c)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}
