### cache-go
This package is implementation of `ARC` and `LRU` cache algorithm.

Simple example:
```go
package main

import (
	"github.com/Sereger/cache-go/arc"
	cacheGC "github.com/Sereger/cache-go/gc"
	"log"
	"time"
)

type SimpleData struct {
	a, b int
}

func main() {
	c1, err := arc.New(128)
	if err != nil {
		log.Fatal(err)
	}
	c2, err := arc.New(128)
	if err != nil {
		log.Fatal(err)
	}

	gc := cacheGC.New(c1, c2)
	defer gc.Close()

	gc.AsyncPurge(8 * time.Second)

	item := SimpleData{1, 2}
	c1.Store("myItem", item)
}
```

So, what's so special about the `ARC` cache?
See example:
```go
cache, _ := New(100)
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

fmt.Printf("keys: %v", keys)
```

Output:
```bash
keys: [1000 999 998 997 996 995 994 993 992 991 990 989 988 987 986 985 984 983 982 981 980 979 978 977 976 26 25 24 23 22 21 20 19 18 17 16 15 14 13 12 11 10 9 8 7 6 5 4 3 2 1]
```
So, we have collection with most usage and recently usage items.

#### Benchmark
```bash
pkg: github.com/Sereger/cache-go/arc
BenchmarkArc/storing-4           3000000               475 ns/op             208 B/op          3 allocs/op
BenchmarkArc/loading-4          30000000                48.1 ns/op             0 B/op          0 allocs/op

pkg: github.com/Sereger/cache-go/lru
BenchmarkLRU/storing-4           5000000               382 ns/op             160 B/op          2 allocs/op
BenchmarkLRU/loading-4          30000000                41.9 ns/op             0 B/op          0 allocs/op
```

