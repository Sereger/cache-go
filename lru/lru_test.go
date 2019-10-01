package lru

import (
	"strconv"
	"testing"
)

func TestLRUCache(t *testing.T) {
	cache := New(100)
	for i := 1000; i > 0; i-- {
		key := strconv.Itoa(i)
		cache.Store(key, i)
		for j := 0; j < i/3; j++ {
			cache.Load(key)
		}
	}
	cache.Purge()

	// after cache.Purge cache have only half of values
	for i := 1; i <= 51; i++ {
		key := strconv.Itoa(i)
		_, ok := cache.Load(key)
		if !ok {
			t.Fatalf("key [%s] not exists\nkeys:%v", key, cache.Keys())
		}
	}
}
