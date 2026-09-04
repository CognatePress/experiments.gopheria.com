package stacks

import (
	"fmt"
	"testing"
)

const maxDepth = 4096

var depths = []int{1, 16, 64, 256, 1024, 4096}

// addrBuf is allocated once so that neither family of benchmarks pays for it.
var addrBuf = make([]uintptr, maxDepth)

// BenchmarkWarmup runs first and its result is discarded. The first benchmark in
// a run is measured under different conditions from the rest; this one absorbs
// that. See /blog/benchstat-or-it-didnt-happen/.
func BenchmarkWarmup(b *testing.B) {
	for b.Loop() {
		Descend(0, addrBuf[:64])
	}
}

// BenchmarkDescend measures the same descent three ways at each depth.
//
//	warm    — a reused goroutine whose stack has already grown; no copies
//	fresh   — a goroutine created for the call, starting at 2 KiB; copies
//	warmctl — identical to warm, and therefore a control: whatever benchstat
//	          reports for it against warm is this harness's own error.
//
// warm and warmctl bracket fresh, so the control spans at least as much drift
// as the comparison it is there to validate.
func BenchmarkDescend(b *testing.B) {
	w := newWorker()
	b.Cleanup(w.stop)
	for _, depth := range depths {
		w.req <- depth
		<-w.done // grow the worker's stack before it is measured at this depth

		b.Run(fmt.Sprintf("goroutine=warm/depth=%d", depth), func(b *testing.B) {
			benchWarm(b, w, depth)
		})
		b.Run(fmt.Sprintf("goroutine=fresh/depth=%d", depth), func(b *testing.B) {
			benchFresh(b, depth)
		})
		b.Run(fmt.Sprintf("goroutine=warmctl/depth=%d", depth), func(b *testing.B) {
			benchWarm(b, w, depth)
		})
	}
}

func benchWarm(b *testing.B, w *worker, depth int) {
	for b.Loop() {
		w.req <- depth
		<-w.done
	}
}

func benchFresh(b *testing.B, depth int) {
	for b.Loop() {
		done := make(chan struct{})
		go func() {
			Descend(0, addrBuf[:depth])
			close(done)
		}()
		<-done
	}
}

type worker struct {
	req  chan int
	done chan struct{}
}

func newWorker() *worker {
	w := &worker{req: make(chan int), done: make(chan struct{})}
	go func() {
		for d := range w.req {
			Descend(0, addrBuf[:d])
			w.done <- struct{}{}
		}
	}()
	return w
}

func (w *worker) stop() { close(w.req) }
