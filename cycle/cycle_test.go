package cycle

import (
	"strconv"
	"testing"
)

func TestLRUCache(t *testing.T) {
	cache := New[int](100)
	for i := 1000; i > 0; i-- {
		key := strconv.Itoa(i)
		cache.Store(key, i)
		for j := 0; j < i/3; j++ {
			cache.Load(key)
		}
	}
	cache.Purge()

	for i := 1; i <= 100; i++ {
		key := strconv.Itoa(i)
		_, ok := cache.Load(key)
		if !ok {
			t.Fatalf("key [%s] not exists", key)
		}
	}
}
