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
