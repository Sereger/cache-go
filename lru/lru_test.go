package lru

import (
	"strconv"
	"testing"
)

func TestLRUCache(t *testing.T) {
	cache, _ := New(100)
	for i := 1000; i > 0; i-- {
		key := strconv.Itoa(i)
		cache.Store(key, i)
		for j := 0; j < i/3; j++ {
			cache.Load(key)
		}
	}
	cache.Purge()

	expectKeys := []string{
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", // resent usage
	}

	for _, key := range expectKeys {
		_, ok := cache.Load(key)
		if !ok {
			t.Fatalf("key [%s] not exists", key)
		}
	}
}
