package cycle

import (
	"strconv"
	"testing"
)

func BenchmarkCycle(b *testing.B) {
	cases := []int{32, 256, 1024, 8192}
	for _, pullSize := range cases {
		b.Run("cache "+strconv.Itoa(pullSize), func(b *testing.B) {
			cache := New(pullSize)

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
		})
	}
}
