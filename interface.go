package cache_go

import "time"

type (
	Cache interface {
		Keys() []string
		Store(key string, val interface{}, opts ...ValueOption)
		Remove(key string)
		Load(key string) (val interface{}, ok bool)
		Purge()
	}

	cell interface {
		SetTTL(ttl time.Duration)
	}

	ValueOption func(c cell)
)

var (
	ValueTTL = func(ttl time.Duration) ValueOption {
		return func(c cell) {
			c.SetTTL(ttl)
		}
	}
)
