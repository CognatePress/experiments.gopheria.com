// Command pool2 runs one shape at one scale and reports what the slot asks for:
// throughput, peak mapped memory, and scheduler latency.
//
//	go build -o /tmp/pool2 ./pool2/main
//	/tmp/pool2 -shape=peritem -items=100000 -work=64
//	/tmp/pool2 -shape=pool    -items=100000 -work=64 -workers=12
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/metrics"
	"time"

	"gopheria.lab/pool2"
)

func main() {
	shape := flag.String("shape", "peritem", "peritem, pool or poolnochan")
	items := flag.Int("items", 100000, "number of work items")
	work := flag.Int("work", 64, "xorshift iterations per item")
	workers := flag.Int("workers", 12, "pool size, ignored by peritem")
	block := flag.Duration("block", 0, "if set, each item sleeps this long instead of computing")
	flag.Parse()

	// Sampled while the run is in flight: peak mapped memory is a peak, and a
	// reading taken afterwards reports whatever the collector left behind.
	stop, peaks := make(chan struct{}), make(chan uint64)
	go samplePeak(stop, peaks)

	before := latencies()
	start := time.Now()
	run(*shape, *items, *work, *workers, *block)
	elapsed := time.Since(start)
	after := latencies()

	close(stop)
	peak := <-peaks

	fmt.Printf("shape        %s\n", *shape)
	fmt.Printf("items        %d\n", *items)
	if *block > 0 {
		fmt.Printf("per item     block %s\n", *block)
	} else {
		fmt.Printf("per item     spin %d\n", *work)
	}
	if *shape != "peritem" {
		fmt.Printf("workers      %d\n", *workers)
	}
	fmt.Printf("elapsed      %s\n", elapsed.Round(time.Microsecond))
	fmt.Printf("throughput   %.3f M items/s\n", float64(*items)/elapsed.Seconds()/1e6)
	fmt.Printf("ns per item  %.1f\n", float64(elapsed.Nanoseconds())/float64(*items))
	fmt.Printf("peak mapped  %.1f MiB\n", float64(peak)/(1<<20))
	fmt.Printf("sched p50    %s\n", pct(before, after, 0.50))
	fmt.Printf("sched p99    %s\n", pct(before, after, 0.99))
	fmt.Printf("sched max    %s\n", pct(before, after, 1.0))
}

func run(shape string, items, work, workers int, block time.Duration) {
	switch shape {
	case "peritem":
		pool2.PerItem(items, work, block)
	case "pool":
		pool2.Pool(items, work, workers, block)
	case "poolnochan":
		pool2.PoolNoChan(items, work, workers, block)
	default:
		fmt.Fprintf(os.Stderr, "unknown shape %q\n", shape)
		os.Exit(2)
	}
}

func samplePeak(stop <-chan struct{}, out chan<- uint64) {
	s := []metrics.Sample{{Name: "/memory/classes/total:bytes"}}
	t := time.NewTicker(2 * time.Millisecond)
	defer t.Stop()
	var peak uint64
	for {
		select {
		case <-stop:
			out <- peak
			return
		case <-t.C:
			metrics.Read(s)
			peak = max(peak, s[0].Value.Uint64())
		}
	}
}

// latencies returns the scheduler latency histogram: how long a goroutine sat
// runnable before a P picked it up. Subtracting two readings gives the
// distribution for the run rather than for the process.
func latencies() *metrics.Float64Histogram {
	s := []metrics.Sample{{Name: "/sched/latencies:seconds"}}
	metrics.Read(s)
	h := s[0].Value.Float64Histogram()
	counts := append([]uint64(nil), h.Counts...)
	buckets := append([]float64(nil), h.Buckets...)
	return &metrics.Float64Histogram{Counts: counts, Buckets: buckets}
}

// pct walks the difference between two histogram readings and returns the
// bucket boundary at which the cumulative count crosses q.
func pct(before, after *metrics.Float64Histogram, q float64) time.Duration {
	var total uint64
	diff := make([]uint64, len(after.Counts))
	for i := range after.Counts {
		diff[i] = after.Counts[i] - before.Counts[i]
		total += diff[i]
	}
	if total == 0 {
		return 0
	}
	target, seen := uint64(float64(total)*q), uint64(0)
	for i, c := range diff {
		seen += c
		if seen >= target {
			return time.Duration(after.Buckets[i+1] * float64(time.Second))
		}
	}
	return time.Duration(after.Buckets[len(after.Buckets)-1] * float64(time.Second))
}
