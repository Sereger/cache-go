package gc

import (
	"time"
)

type (
	cache interface {
		Purge()
	}
	GC struct {
		list   []cache
		closed chan struct{}
	}
)

func New(list ...cache) *GC {
	return &GC{list: list, closed: make(chan struct{})}
}

func (gc *GC) AsyncPurge(timeout time.Duration) {
	go func() {
		tik := time.NewTicker(timeout)
		defer tik.Stop()

		for {
			select {
			case <-tik.C:
				gc.Purge()
			case <-gc.closed:
				return
			}
		}
	}()
}

func (gc *GC) Close() error {
	close(gc.closed)
	return nil
}

func (gc *GC) Purge() {
	for _, item := range gc.list {
		item.Purge()
	}
}
