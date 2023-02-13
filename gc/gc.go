package gc

import (
	"sync/atomic"
	"time"
)

type (
	cache interface {
		Purge()
	}
	GC struct {
		list   []cache
		closed uint32
	}
)

func New(list ...cache) *GC {
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
