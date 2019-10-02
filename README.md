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
	c1 := arc.New(128)
	c2 := arc.New(128)

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

fmt.Printf("keys: %v", keys)
```

Output:
```bash
keys: [1000 999 998 997 996 995 994 993 992 991 990 989 988 987 986 985 16 15 14 13 12 11 10 9 8 7 6 5 4 3 2 1]
```
So, we have collection with most usage and recently usage items.

#### Benchmark
This results on MacBook Pro (15-inch, Mid 2015)
```bash
pkg: github.com/Sereger/cache-go/arc
BenchmarkArc/storing-8           1957212               631 ns/op             104 B/op          3 allocs/op
BenchmarkArc/loading-8          20467015              59.6 ns/op               7 B/op          0 allocs/op

pkg: github.com/Sereger/cache-go/lru
BenchmarkCycle/storing-8         2313512               475 ns/op              81 B/op          3 allocs/op
BenchmarkCycle/loading-8        20757928              58.6 ns/op               7 B/op          0 allocs/op

pkg: github.com/Sereger/cache-go/cycle
BenchmarkLRU/storing-8           1671874               720 ns/op             293 B/op          3 allocs/op
BenchmarkLRU/loading-8          20843748              58.4 ns/op               7 B/op          0 allocs/op
```

