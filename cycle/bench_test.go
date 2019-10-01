package cycle

import (
	"strconv"
	"testing"
)

func BenchmarkCycle(b *testing.B) {
	cache := New(b.N/4 + 8)

	b.Run("storing", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			var i int
			for pb.Next() {
				key := strconv.Itoa(i)
				cache.Store(key, i)
				i++
			}
		})
	})
	b.Run("loading", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			var i int
			for pb.Next() {
				key := strconv.Itoa(i)
				cache.Load(key)
				i++
			}
		})
	})
}
