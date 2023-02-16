package arc

import (
	"strconv"
	"testing"
)

func TestARCCache(t *testing.T) {
	cache := New[string, int](100)
	for i := 1000; i > 0; i-- {
		key := strconv.Itoa(i)
		cache.Store(key, i)
		for j := 0; j < i/3; j++ {
			cache.Load(key)
		}
		if i%3 == 0 {
			cache.Remove(key)
		}
	}
	cache.Purge()

	expectKeys := []string{
		"1000", "998", "997", "995", "994", "992", "991", // most usage
		"1", "2", "4", "5", "7", "8", "10", // resent usage
	}

	for _, key := range expectKeys {
		_, ok := cache.Load(key)
		if !ok {
			t.Fatalf("key [%s] not exists", key)
		}
	}
}
func TestARCCache_Inc(t *testing.T) {
	cache := New[string, int](100)
	key := "testKey"
	for i := 1; i < 1000; i++ {
		v := cache.Atomic(key, func(v int, _ bool) int {
			v++
			return v
		})

		if v != i {
			t.Fatalf("v != i (%d != %d)", v, i)
		}
	}
}
