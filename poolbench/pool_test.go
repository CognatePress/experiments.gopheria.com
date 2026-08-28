package poolbench

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

type buf struct{ b [4096]byte }

// keep makes the buffer outlive the call, which is the whole reason a pool
// would exist. Without it the compiler stack-allocates the "unpooled" case and
// the comparison measures nothing.
var keep *buf

//go:noinline
func use(x *buf) {
	x.b[0]++
	keep = x
}

func BenchmarkNoPool(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		use(new(buf))
	}
}

func BenchmarkPool(b *testing.B) {
	var misses int64
	p := &sync.Pool{New: func() any { misses++; return new(buf) }}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := p.Get().(*buf)
		use(x)
		p.Put(x)
	}
	b.ReportMetric(float64(misses)/float64(b.N)*100, "%miss")
}

// Same pool, with a GC every millisecond underneath it — roughly what a busy
// service with a small heap looks like.
func BenchmarkPoolUnderGC(b *testing.B) {
	var misses int64
	p := &sync.Pool{New: func() any { misses++; return new(buf) }}

	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		t := time.NewTicker(time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				runtime.GC()
			}
		}
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := p.Get().(*buf)
		use(x)
		p.Put(x)
	}
	b.StopTimer()
	close(stop)
	wg.Wait()
	b.ReportMetric(float64(misses)/float64(b.N)*100, "%miss")
}
