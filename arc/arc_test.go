package arc

import (
	"strconv"
	"testing"
)

func TestARCCache(t *testing.T) {
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
		"1000", "999", "998", "997", "996", "995", "994", "993", "992", "991", "990", // most usage
		"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", // resent usage
	}

	for _, key := range expectKeys {
		_, ok := cache.Load(key)
		if !ok {
			t.Fatalf("key [%s] not exists", key)
		}
	}
}
