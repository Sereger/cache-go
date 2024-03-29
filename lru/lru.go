package lru

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"

	cacheGo "github.com/Sereger/cache-go/v2"
)

type (
	Cache[K comparable, T any] struct {
		lock   sync.RWMutex
		keyMap map[K]int
		buff   []*cell[K, T]
		idx    int
	}

	cell[K comparable, T any] struct {
		key      K
		value    T
		removed  uint32
		lastRead int64
		expired  time.Time
	}
)

func New[K comparable, T any](n int) *Cache[K, T] {
	if n < 8 {
		n = 8
	}
	return &Cache[K, T]{
		keyMap: make(map[K]int),
		buff:   make([]*cell[K, T], n),
	}
}

func (c *Cache[K, T]) Keys() []K {
	c.lock.RLock()
	defer c.lock.RUnlock()

	result := make([]K, 0, c.idx)
	for i := 0; i < c.idx; i++ {
		v := c.buff[i]
		if v.isRemoved() {
			continue
		}

		if !v.expired.IsZero() && v.expired.Before(time.Now()) {
			v.markRemoved()
			continue
		}

		result = append(result, c.buff[i].key)
	}

	return result
}

func (c *Cache[K, T]) Values() []T {
	c.lock.RLock()
	defer c.lock.RUnlock()

	result := make([]T, 0, c.idx)
	for i := 0; i < c.idx; i++ {
		v := c.buff[i]
		if v.isRemoved() {
			continue
		}

		if !v.expired.IsZero() && v.expired.Before(time.Now()) {
			v.markRemoved()
			continue
		}

		result = append(result, c.buff[i].value)
	}

	return result
}

func (c *Cache[K, T]) Store(key K, val T, opts ...cacheGo.ValueOption) {
	c.lock.Lock()
	defer c.lock.Unlock()

	cell := &cell[K, T]{key: key, value: val}
	for _, opt := range opts {
		opt(cell)
	}

	i, ok := c.keyMap[key]
	if ok {
		c.buff[i] = cell
		return
	}

	v := c.buff[c.idx]
	if v != nil {
		delete(c.keyMap, v.key)
	}
	c.buff[c.idx] = cell
	c.keyMap[key] = c.idx
	c.idx++
	if c.idx == len(c.buff) {
		c.purge()
	}
}
func (c *Cache[K, T]) Remove(key K) {
	v, ok := c.loadActCell(key)
	if !ok {
		return
	}
	v.markRemoved()
}

func (c *Cache[K, T]) loadActCell(key K) (*cell[K, T], bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	idx, ok := c.keyMap[key]
	if !ok {
		return nil, false
	}
	v := c.buff[idx]

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
	atomic.StoreInt64(&v.lastRead, time.Now().Unix())
	return v.value, true
}

func (c *cell[K, T]) markRemoved() {
	atomic.StoreUint32(&c.removed, 1)
}

func (c *cell[K, T]) isRemoved() bool {
	return atomic.LoadUint32(&c.removed) == 1
}

func (c *Cache[K, T]) Purge() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.purge()
}

func (c *Cache[K, T]) purge() {
	c.idx = 0
	moment := time.Now()
	sort.Slice(c.buff, func(i, j int) bool {
		v1, v2 := c.buff[i], c.buff[j]

		if v1 == nil && v2 != nil {
			return false
		} else if v2 == nil && v1 != nil {
			return true
		} else if v2 == nil && v1 == nil {
			return true
		}

		rm1, rm2 := v1.isRemoved(), v2.isRemoved()
		if !v1.expired.IsZero() && !rm1 && v1.expired.Before(moment) {
			v1.markRemoved()
			rm1 = true
		}

		if !v2.expired.IsZero() && !rm2 && v2.expired.Before(moment) {
			v2.markRemoved()
			rm2 = true
		}

		if rm1 && !rm2 {
			return false
		} else if !rm1 && rm2 {
			return true
		} else if rm1 && rm2 {
			return true
		}

		return v1.lastRead < v2.lastRead
	})

	idx := len(c.buff) - 1
	for {
		if c.buff[idx] != nil && !c.buff[idx].isRemoved() {
			break
		}
		idx--
	}

	if idx == len(c.buff)-1 {
		idx = 0 // nolint: ineffassign
	}
	for i, v := range c.buff {
		if v.removed == 1 {
			delete(c.keyMap, v.key)
			continue
		}
		c.keyMap[v.key] = i
	}
}

func (c *cell[K, T]) SetTTL(ttl time.Duration) {
	c.expired = time.Now().Add(ttl)
}
