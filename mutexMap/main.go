package mutexMap

import (
	cacheGo "github.com/Sereger/cache-go"
	"sync"
	"sync/atomic"
	"time"
)

type (
	Cache struct {
		lock   sync.RWMutex
		values map[string]*cell
	}

	cell struct {
		value   interface{}
		removed uint32
		expired time.Time
	}
)

func New() *Cache {
	return &Cache{
		values: make(map[string]*cell),
	}
}

func (c *Cache) Keys() []string {
	result := make([]string, 0, len(c.values))
	for key := range c.values {
		result = append(result, key)
	}

	return result
}

func (c *Cache) Store(key string, val interface{}, opts ...cacheGo.ValueOption) {
	c.lock.Lock()
	defer c.lock.Unlock()

	cell := &cell{value: val}
	for _, opt := range opts {
		opt(cell)
	}

	c.values[key] = cell
}
func (c *Cache) Remove(key string) {
	v, ok := c.loadActCell(key)
	if !ok {
		return
	}
	v.markRemoved()
}

func (c *Cache) loadActCell(key string) (*cell, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	v, ok := c.values[key]
	if !ok {
		return nil, false
	}

	if v.isRemoved() {
		return nil, false
	}

	if !v.expired.IsZero() && v.expired.Before(time.Now()) {
		v.markRemoved()
		return nil, false
	}

	return v, true
}

func (c *Cache) Load(key string) (interface{}, bool) {
	v, ok := c.loadActCell(key)
	if !ok {
		return nil, false
	}
	return v.value, true
}

func (c *cell) markRemoved() {
	atomic.StoreUint32(&c.removed, 1)
}

func (c *cell) isRemoved() bool {
	return atomic.LoadUint32(&c.removed) == 1
}
func (c *Cache) Purge() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.purge()
}

func (c *Cache) purge() {
	delKeys := make([]string, 0, len(c.values))

	for k, v := range c.values {
		if v.isRemoved() {
			delKeys = append(delKeys, k)
		}
	}

	for _, key := range delKeys {
		delete(c.values, key)
	}
}

func (c *cell) SetTTL(ttl time.Duration) {
	c.expired = time.Now().Add(ttl)
}
