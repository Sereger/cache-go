package cycle

import (
	"fmt"
	"strconv"
	"testing"
)

func BenchmarkCycle(b *testing.B) {
	cases := []int{32, 256, 1024, 8192}
	for _, pullSize := range cases {
		name := fmt.Sprintf("size %4d", pullSize)
		b.Run(name, func(b *testing.B) {
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

			b.Run("storing/removing", func(b *testing.B) {
				b.RunParallel(func(pb *testing.PB) {
					var (
						i       int
						lastKey string
					)
					for pb.Next() {
						key := strconv.Itoa(i)
						cache.Store(key, i)
						if i%3 == 0 {
							cache.Remove(lastKey)
						}
						i++
						lastKey = key
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
