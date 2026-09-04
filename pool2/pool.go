// Package pool2 compares a fixed worker pool against one goroutine per item.
//
// The worker pool is imported from languages where a thread is expensive. This
// measures what it costs and saves in Go, at three orders of magnitude, with
// the per-item work varied between something that runs and something that
// blocks — because those two put the answer in different places.
package pool2

import (
	"sync"
	"sync/atomic"
	"time"
)

// Sink keeps the per-item results reachable so the compiler cannot delete the
// work being measured.
var Sink atomic.Uint64

// spin is compute-bound per-item work. It costs about the same wherever it
// runs, so it isolates scheduling from the work itself.
//
//go:noinline
func spin(n int) uint64 {
	x := uint64(0x9e3779b97f4a7c15)
	for i := 0; i < n; i++ {
		x ^= x << 13
		x ^= x >> 7
		x ^= x << 17
	}
	return x
}

// PerItem starts one goroutine per item. This is the shape people are told not
// to write, and the one Go's runtime is built for.
func PerItem(items, work int, block time.Duration) {
	var wg sync.WaitGroup
	wg.Add(items)
	for i := 0; i < items; i++ {
		go func() {
			defer wg.Done()
			do(work, block)
		}()
	}
	wg.Wait()
}

// do is the per-item work: either compute, or block. A blocking item is what a
// pool is usually defending against, and the two cases put the answer in
// different places.
func do(work int, block time.Duration) {
	if block > 0 {
		time.Sleep(block)
		return
	}
	Sink.Add(spin(work))
}

// Pool runs the same items through a fixed number of workers reading from a
// channel. The channel is buffered at the worker count, which is the shape most
// worker-pool code has.
func Pool(items, work, workers int, block time.Duration) {
	ch := make(chan int, workers)
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for range ch {
				do(work, block)
			}
		}()
	}
	for i := 0; i < items; i++ {
		ch <- i
	}
	close(ch)
	wg.Wait()
}

// PoolNoChan is the pool without the channel: workers claim items from a shared
// counter instead. It separates the cost of bounding concurrency from the cost
// of the queue used to bound it, which is the confound in most published
// comparisons.
func PoolNoChan(items, work, workers int, block time.Duration) {
	var next atomic.Int64
	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		go func() {
			defer wg.Done()
			for {
				i := next.Add(1)
				if i > int64(items) {
					return
				}
				do(work, block)
			}
		}()
	}
	wg.Wait()
}
