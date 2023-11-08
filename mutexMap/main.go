package mutexMap // nolint: stylecheck

import (
	"sync"
	"sync/atomic"
	"time"

	cacheGo "github.com/Sereger/cache-go/v2"
)

type (
	Cache[K comparable, T any] struct {
		lock   sync.RWMutex
		values map[K]*cell[T]
	}

	cell[T any] struct {
		value   T
		removed uint32
		expired time.Time
	}
)

func New[K comparable, T any]() *Cache[K, T] {
	return &Cache[K, T]{
		values: make(map[K]*cell[T]),
	}
}

func (c *Cache[K, T]) Keys() []K {
	result := make([]K, 0, len(c.values))
	for key := range c.values {
		result = append(result, key)
	}

	return result
}

func (c *Cache[K, T]) Values() []T {
	c.lock.RLock()
	defer c.lock.RUnlock()

	result := make([]T, 0, len(c.values))
	for _, v := range c.values {
		result = append(result, v.value)
	}

	return result
}

func (c *Cache[K, T]) Store(key K, val T, opts ...cacheGo.ValueOption) {
	c.lock.Lock()
	defer c.lock.Unlock()

	cell := &cell[T]{value: val}
	for _, opt := range opts {
		opt(cell)
	}

	c.values[key] = cell
}
func (c *Cache[K, T]) Remove(key K) {
	v, ok := c.loadActCell(key)
	if !ok {
		return
	}
	v.markRemoved()
}

func (c *Cache[K, T]) loadActCell(key K) (*cell[T], bool) {
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

func (c *Cache[K, T]) Load(key K) (T, bool) {
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
func (c *Cache[K, T]) Purge() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.purge()
}

func (c *Cache[K, T]) purge() {
	delKeys := make([]K, 0, len(c.values))

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
