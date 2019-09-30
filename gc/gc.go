package gc

import (
	"github.com/Sereger/cache-go"
	"sync/atomic"
	"time"
)

type GC struct {
	list   []cache_go.Cache
	closed uint32
}

func New(list ...cache_go.Cache) *GC {
	return &GC{list: list}
}

func (gc *GC) AsyncPurge(timeout time.Duration) {
	for atomic.LoadUint32(&gc.closed) == 0 {
		time.Sleep(timeout)
		gc.Purge()
	}
}

func (gc *GC) Close() error {
	atomic.StoreUint32(&gc.closed, 1)
	return nil
}

func (gc *GC) Purge() {
	for _, item := range gc.list {
		item.Purge()
	}
}
