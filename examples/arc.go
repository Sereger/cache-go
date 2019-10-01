package main

import (
	"github.com/Sereger/cache-go/arc"
	cacheGC "github.com/Sereger/cache-go/gc"
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
