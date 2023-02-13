package mutexMap

import (
	cacheGo "github.com/Sereger/cache-go/v2"
	"sync"
	"sync/atomic"
	"time"
)

type (
	Cache[T any] struct {
		lock   sync.RWMutex
		values map[string]*cell[T]
	}

	cell[T any] struct {
		value   T
		removed uint32
		expired time.Time
	}
)

func New[T any]() *Cache[T] {
	return &Cache[T]{
		values: make(map[string]*cell[T]),
	}
}

func (c *Cache[T]) Keys() []string {
	result := make([]string, 0, len(c.values))
	for key := range c.values {
		result = append(result, key)
	}

	return result
}

func (c *Cache[T]) Store(key string, val T, opts ...cacheGo.ValueOption) {
	c.lock.Lock()
	defer c.lock.Unlock()

	cell := &cell[T]{value: val}
	for _, opt := range opts {
		opt(cell)
	}

	c.values[key] = cell
}
func (c *Cache[T]) Remove(key string) {
	v, ok := c.loadActCell(key)
	if !ok {
		return
	}
	v.markRemoved()
}

func (c *Cache[T]) loadActCell(key string) (*cell[T], bool) {
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

func (c *Cache[T]) Load(key string) (T, bool) {
	v, ok := c.loadActCell(key)
	if !ok {
		var result T
		return result, false
	}
	return v.value, true
}

func (c *cell[T]) markRemoved() {
	atomic.StoreUint32(&c.removed, 1)
}

func (c *cell[T]) isRemoved() bool {
	return atomic.LoadUint32(&c.removed) == 1
}
func (c *Cache[T]) Purge() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.purge()
}

func (c *Cache[T]) purge() {
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

func (c *cell[T]) SetTTL(ttl time.Duration) {
	c.expired = time.Now().Add(ttl)
}
