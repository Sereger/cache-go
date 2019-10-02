package main

import (
	"fmt"
	"github.com/Sereger/cache-go/arc"
	"sort"
	"strconv"
)

func main() {
	cache := arc.New(32)
	for i := 1000; i > 0; i-- {
		key := strconv.Itoa(i)
		cache.Store(key, i)
		for j := 0; j < i/3; j++ {
			cache.Load(key) // so, the largest key will have more readings
		}
	}

	keys := cache.Keys()

	// For pretty printing
	sort.Slice(keys, func(i, j int) bool {
		v1, _ := strconv.Atoi(keys[i])
		v2, _ := strconv.Atoi(keys[j])
		return v1 > v2
	})

	fmt.Printf("keys: %v\n", keys)
}
