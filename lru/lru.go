package lru

import (
	cacheGo "github.com/Sereger/cache-go"
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
		rmCtn  uint32
	}

	cell struct {
		key     string
		value   interface{}
		removed uint32
		expired time.Time
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
	atomic.AddUint32(&c.rmCtn, 1)
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
		atomic.AddUint32(&c.rmCtn, 1)
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
	c.idx = 0
	if c.rmCtn == 0 {
		return
	}
	c.rmCtn = 0
	moment := time.Now()
	var subIdx int
	for i, v := range c.buff {
		if v == nil {
			c.idx = i
			break
		}

		var shouldDel bool
		if v.isRemoved() {
			shouldDel = true
		} else if !v.expired.IsZero() && v.expired.Before(moment) {
			shouldDel = true
		}

		if !shouldDel {
			continue
		}

		if subIdx == 0 {
			subIdx = i + 1
		}
		for subIdx < len(c.buff) {
			v2 := c.buff[subIdx]
			if v2 == nil {
				return
			}
			if v2.isRemoved() {
				subIdx++
				continue
			} else if !v2.expired.IsZero() && v2.expired.Before(moment) {
				v2.markRemoved()
				subIdx++
				continue
			}

			c.buff[i] = v2
			c.buff[subIdx] = v
			c.keyMap[v2.key] = i
			c.keyMap[v.key] = subIdx
			subIdx++
			break
		}
		if subIdx == len(c.buff) {
			return
		}
	}
}

func (c *cell) SetTTL(ttl time.Duration) {
	c.expired = time.Now().Add(ttl)
}
