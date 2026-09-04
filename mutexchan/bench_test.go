package mutexchan

import (
	"flag"
	"fmt"
	"sync"
	"testing"
)

// spinIters is the number of xorshift iterations executed inside the critical
// section. The sweep is run twice — once at 0 and once at a value priced by
// BenchmarkSpin — because an empty critical section and a busy one put the
// crossover in different places, and only one of them is the case anybody
// actually has.
//
// The flag is -spin rather than -work because `go test` owns -work.
var spinIters = flag.Int("spin", 0, "xorshift iterations inside the critical section")

// goroutines is the contention axis. It runs past GOMAXPROCS on purpose: the
// interesting region for a channel-owned counter starts where there are more
// senders than cores.
var goroutines = []int{1, 2, 4, 8, 16, 32, 64}

// BenchmarkSpin prices the in-section work on its own, so the sweep's work
// level can be quoted in nanoseconds rather than in iterations.
func BenchmarkSpin(b *testing.B) {
	for _, n := range []int{0, 8, 16, 32, 64, 128} {
		b.Run(fmt.Sprintf("iters=%d", n), func(b *testing.B) {
			var acc uint64
			for b.Loop() {
				acc += spin(n)
			}
			sink.Store(acc)
		})
	}
}

// BenchmarkWarmup runs before BenchmarkCounter and its result is discarded.
// Slot 02's spread runs showed the first benchmark in a process pays a penalty
// the rest do not; this is where that penalty is spent.
func BenchmarkWarmup(b *testing.B) {
	c := NewMutex(0)
	inc := c.Handle()
	for b.Loop() {
		inc()
	}
	if c.Close() == 0 {
		b.Fatal("counter never advanced")
	}
}

// BenchmarkCounter is the sweep: five implementations against seven contention
// levels, one shared counter each, b.N increments in total however many
// goroutines share them out.
//
//	go test -run=^$ -bench='Warmup|Counter' -benchmem -count=10 ./mutexchan |
//	  grep -v Warmup | benchstat -row /goroutines -col /impl -
func BenchmarkCounter(b *testing.B) {
	impls := []struct {
		name string
		new  func(work int) Counter
		// emptyOnly marks an implementation that has no critical section to put
		// work into, and so appears only in the work=0 table.
		emptyOnly bool
	}{
		{name: "impl=mutex", new: NewMutex},
		// mutex-control is a second copy of the baseline. Its true delta
		// against impl=mutex is zero, so whatever benchstat reports for it is
		// this session's error bar — the control pair from slot 02.
		{name: "impl=mutex-control", new: NewMutex},
		{name: "impl=atomic", new: func(int) Counter { return NewAtomic() }, emptyOnly: true},
		{name: "impl=chan-unbuf", new: func(w int) Counter { return NewChan(0, w) }},
		{name: "impl=chan-buf1024", new: func(w int) Counter { return NewChan(1024, w) }},
		{name: "impl=chan-ack", new: NewAck},
	}

	for _, im := range impls {
		if im.emptyOnly && *spinIters > 0 {
			continue
		}
		b.Run(im.name, func(b *testing.B) {
			for _, n := range goroutines {
				b.Run(fmt.Sprintf("goroutines=%d", n), func(b *testing.B) {
					sweep(b, n, im.new(*spinIters))
				})
			}
		})
	}
}

// sweep shares b.N increments evenly across n goroutines and times the lot.
//
// The goroutines are created inside the timed region, so allocs/op carries n
// stacks amortised over b.N operations rather than a per-operation figure. At
// the b.N values this benchmark reaches that is a rounding error, and the
// alternative — a barrier that parks every goroutine until the last one is
// ready — measures the barrier at the low contention levels.
func sweep(b *testing.B, n int, c Counter) {
	handles := make([]func(), n)
	for i := range handles {
		handles[i] = c.Handle()
	}

	per, extra := b.N/n, b.N%n
	var wg sync.WaitGroup
	wg.Add(n)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < n; i++ {
		k := per
		if i < extra {
			k++
		}
		go func(inc func(), k int) {
			defer wg.Done()
			for j := 0; j < k; j++ {
				inc()
			}
		}(handles[i], k)
	}
	wg.Wait()
	b.StopTimer()

	// The count is the correctness check. An asynchronous implementation that
	// returned early would show up here as a short total rather than as a fast
	// benchmark.
	if got := c.Close(); got != int64(b.N) {
		b.Fatalf("counter = %d, want %d", got, b.N)
	}
}
