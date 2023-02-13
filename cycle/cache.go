package cycle

import (
	cacheGo "github.com/Sereger/cache-go/v2"
	"sync"
	"sync/atomic"
	"time"
)

type (
	Cache[T any] struct {
		lock   sync.RWMutex
		keyMap map[string]int
		buff   []*cell[T]
		idx    int
		rmCtn  uint32
	}

	cell[T any] struct {
		key     string
		value   T
		removed uint32
		expired time.Time
	}
)

func New[T any](n int) *Cache[T] {
	if n < 8 {
		n = 8
	}
	return &Cache[T]{
		keyMap: make(map[string]int),
		buff:   make([]*cell[T], n),
	}
}

func (c *Cache[T]) Keys() []string {
	result := make([]string, 0, len(c.buff))
	for key := range c.keyMap {
		result = append(result, key)
	}

	return result
}

func (c *Cache[T]) Store(key string, val T, opts ...cacheGo.ValueOption) {
	c.lock.Lock()
	defer c.lock.Unlock()

	cell := &cell[T]{key: key, value: val}
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
func (c *Cache[T]) Remove(key string) {
	v, ok := c.loadActCell(key)
	if !ok {
		return
	}
	atomic.AddUint32(&c.rmCtn, 1)
	v.markRemoved()
}

func (c *Cache[T]) loadActCell(key string) (*cell[T], bool) {
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

func (c *cell[T]) SetTTL(ttl time.Duration) {
	c.expired = time.Now().Add(ttl)
}
