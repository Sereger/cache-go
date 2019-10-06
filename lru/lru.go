package lru

import (
	cacheGo "github.com/Sereger/cache-go"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type (
	Cache struct {
		lock   sync.RWMutex
		keyMap map[string]int
		buff   []*cell
		idx    int
	}

	cell struct {
		key      string
		value    interface{}
		removed  uint32
		lastRead int64
		expired  time.Time
	}
)

func New(n int) *Cache {
	if n < 8 {
		n = 8
	}
	return &Cache{
		keyMap: make(map[string]int),
		buff:   make([]*cell, n),
	}
}

func (c *Cache) Keys() []string {
	result := make([]string, 0, len(c.buff))
	for key := range c.keyMap {
		result = append(result, key)
	}

	return result
}

func (c *Cache) Store(key string, val interface{}, opts ...cacheGo.ValueOption) {
	c.lock.Lock()
	defer c.lock.Unlock()

	cell := &cell{key: key, value: val}
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

func (c *Cache) Load(key string) (interface{}, bool) {
	v, ok := c.loadActCell(key)
	if !ok {
		return nil, false
	}
	atomic.StoreInt64(&v.lastRead, time.Now().Unix())
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
		idx = 0
	}
	for i, v := range c.buff {
		if v.removed == 1 {
			delete(c.keyMap, v.key)
			continue
		}
		c.keyMap[v.key] = i
	}
}

func (c *cell) SetTTL(ttl time.Duration) {
	c.expired = time.Now().Add(ttl)
}
