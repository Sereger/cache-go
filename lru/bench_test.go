package lru

import (
	"strconv"
	"testing"
)

func BenchmarkLRU(b *testing.B) {
	cache, err := New(b.N + 8)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("storing", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			var i int
			for pb.Next() {
				key := strconv.Itoa(i)
				cache.Store(key, i)
			}
		})
	})
	b.Run("loading", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			var i int
			for pb.Next() {
				key := strconv.Itoa(i)
				cache.Load(key)
			}
		})
	})
}
