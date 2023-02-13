### cache-go

This package is implementation of `ARC` and `LRU` cache algorithm.

Simple example:

```go
package main

import (
	"github.com/Sereger/cache-go/v2/arc"
	cacheGC "github.com/Sereger/cache-go/v2/gc"
	"time"
)

type SimpleData struct {
	a, b int
}

func main() {
	c1 := arc.New[SimpleData](128)
	c2 := arc.New[int](128)

	gc := cacheGC.New(c1, c2)
	defer gc.Close()

	gc.AsyncPurge(8 * time.Second)

	item := SimpleData{a: 1, b: 2}
	c1.Store("myItem", item)
}
```

So, what's so special about the `ARC` cache? See example:

```go
cache := arc.New[int](32)
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

This results on MacBook Pro (13-inch, 2019)
```bash
pkg: github.com/Sereger/cache-go/arc
BenchmarkArc/size___32/storing-8         2901594               415 ns/op             104 B/op          3 allocs/op
BenchmarkArc/size___32/loading-8        25883810                46.9 ns/op             7 B/op          0 allocs/op
BenchmarkArc/size__256/storing-8         2731152               440 ns/op              97 B/op          3 allocs/op
BenchmarkArc/size__256/loading-8        25534323                47.6 ns/op             7 B/op          0 allocs/op
BenchmarkArc/size_1024/storing-8         2560555               473 ns/op              96 B/op          3 allocs/op
BenchmarkArc/size_1024/loading-8        25559084                48.7 ns/op             7 B/op          0 allocs/op
BenchmarkArc/size_8192/storing-8         2480404               497 ns/op              96 B/op          3 allocs/op
BenchmarkArc/size_8192/loading-8        24895506                51.2 ns/op             7 B/op          0 allocs/op

pkg: github.com/Sereger/cache-go/cycle
BenchmarkCycle/size___32/storing-8       4127497               291 ns/op              82 B/op          3 allocs/op
BenchmarkCycle/size___32/loading-8      26939912                46.9 ns/op             7 B/op          0 allocs/op
BenchmarkCycle/size__256/storing-8       3946203               299 ns/op              81 B/op          3 allocs/op
BenchmarkCycle/size__256/loading-8      26589692                47.1 ns/op             7 B/op          0 allocs/op
BenchmarkCycle/size_1024/storing-8       3873979               316 ns/op              80 B/op          3 allocs/op
BenchmarkCycle/size_1024/loading-8      26339017                48.7 ns/op             7 B/op          0 allocs/op
BenchmarkCycle/size_8192/storing-8       3427587               359 ns/op              80 B/op          3 allocs/op
BenchmarkCycle/size_8192/loading-8      25744363                50.8 ns/op             7 B/op          0 allocs/op

pkg: github.com/Sereger/cache-go/lru
BenchmarkLRU/size___32/storing-8         3237692               360 ns/op             100 B/op          3 allocs/op
BenchmarkLRU/size___32/loading-8        28123472                44.8 ns/op             7 B/op          0 allocs/op
BenchmarkLRU/size__256/storing-8         3393638               356 ns/op              97 B/op          3 allocs/op
BenchmarkLRU/size__256/loading-8        28250967                45.2 ns/op             7 B/op          0 allocs/op
BenchmarkLRU/size_1024/storing-8         3317509               373 ns/op              96 B/op          3 allocs/op
BenchmarkLRU/size_1024/loading-8        28495053                46.3 ns/op             7 B/op          0 allocs/op
BenchmarkLRU/size_8192/storing-8         3197606               402 ns/op              96 B/op          3 allocs/op
BenchmarkLRU/size_8192/loading-8        25213560                50.3 ns/op             7 B/op          0 allocs/op
```

