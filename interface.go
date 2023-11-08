package cache_go // nolint: stylecheck

import "time"

type (
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
