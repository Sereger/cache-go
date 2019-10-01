package lru

import (
	cacheGo "github.com/Sereger/cache-go"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type (
	LRUCache struct {
		lock   sync.RWMutex
		keyMap map[string]int
		buff   []*cell
		idx    int
		age    int
	}

	cell struct {
		key     string
		value   interface{}
		age     int
		removed uint32
		expired time.Time
	}
)

func New(n int) *LRUCache {
	if n < 8 {
		n = 8
	}
	return &LRUCache{
		keyMap: make(map[string]int),
		buff:   make([]*cell, n),
		age:    1,
	}
}

func (c *LRUCache) Keys() []string {
	result := make([]string, 0, len(c.buff))
	for key := range c.keyMap {
		result = append(result, key)
	}

	return result
}

func (c *LRUCache) Store(key string, val interface{}, opts ...cacheGo.ValueOption) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.age++
	v := &cell{value: val, key: key, age: c.age}

	for _, opt := range opts {
		opt(v)
	}

	i, ok := c.keyMap[key]
	if ok {
		c.buff[i] = v
		return
	}

	if c.idx == len(c.buff) {
		c.purge()
	}
	c.keyMap[key] = c.idx
	c.buff[c.idx] = v
	c.idx++
}

func (c *LRUCache) Remove(key string) {
	v, ok := c.loadActCell(key)
	if !ok {
		return
	}

	v.markRemoved()
}

func (c *LRUCache) loadActCell(key string) (*cell, bool) {
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

func (c *LRUCache) Load(key string) (interface{}, bool) {
	v, ok := c.loadActCell(key)
	if !ok {
		return nil, false
	}

	return v.value, true
}

func (c *LRUCache) Purge() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.purge()
}

func (c *LRUCache) purge() {
	if c.idx <= (len(c.buff) / 3 * 2) {
		return
	}

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

		return v1.age > v2.age
	})

	lastIdx := len(c.buff)/2 + 1
	v := c.buff[lastIdx-1]
	for v == nil || v.isRemoved() {
		lastIdx--
		if lastIdx < 1 {
			lastIdx = 1
			break
		}
		v = c.buff[lastIdx-1]
	}

	c.idx = lastIdx
	newMap := make(map[string]int, lastIdx)

	for i := 0; i < lastIdx; i++ {
		v := c.buff[i]
		newMap[v.key] = i
	}
	c.keyMap = newMap
}

func (c *cell) markRemoved() {
	atomic.StoreUint32(&c.removed, 1)
}

func (c *cell) isRemoved() bool {
	return atomic.LoadUint32(&c.removed) == 1
}

func (c *cell) SetTTL(ttl time.Duration) {
	c.expired = time.Now().Add(ttl)
}
