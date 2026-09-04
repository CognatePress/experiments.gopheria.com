// Command heapprof produces two heap profiles with opposite shapes.
//
// churn allocates fast and keeps almost nothing; retain allocates slowly and
// keeps everything. Their alloc_space and inuse_space profiles disagree about
// which function matters, which is the point: the two sample types answer
// different questions, and reading the wrong one sends you to the wrong
// function.
//
//	go build -o /tmp/heapprof ./heapprof
//	/tmp/heapprof -mode=churn  -out=/tmp/churn.pprof
//	/tmp/heapprof -mode=retain -out=/tmp/retain.pprof
//	go tool pprof -top -nodecount=5 -sample_index=alloc_space /tmp/churn.pprof
//	go tool pprof -top -nodecount=5 -sample_index=inuse_space /tmp/churn.pprof
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// Two allocating functions with distinct names, so the profile's per-function
// view has something to disagree about. Neither is called by the other.

// churnBuffers allocates and immediately abandons. Nothing it produces is
// reachable by the time the profile is written.
//
//go:noinline
func churnBuffers(n int) int {
	total := 0
	for i := 0; i < n; i++ {
		b := make([]byte, 4096)
		b[0] = byte(i)
		total += len(b)
		abandoned = b
	}
	return total
}

// retainBuffers allocates far less and keeps every byte of it.
//
//go:noinline
func retainBuffers(n int) int {
	for i := 0; i < n; i++ {
		b := make([]byte, 4096)
		b[0] = byte(i)
		held = append(held, b)
	}
	return len(held)
}

var (
	abandoned []byte
	held      [][]byte
)

func main() {
	mode := flag.String("mode", "churn", "churn or retain")
	out := flag.String("out", "/tmp/heap.pprof", "where to write the heap profile")
	rate := flag.Int("rate", 0, "runtime.MemProfileRate; 0 leaves the default of 512 KiB")
	flag.Parse()

	// MemProfileRate has to be set before the first allocation being sampled,
	// which in practice means before the workload starts. A heap profile is a
	// sample, and this is the knob that decides how coarse a one.
	if *rate > 0 {
		runtime.MemProfileRate = *rate
	}

	start := time.Now()
	var n int
	switch *mode {
	case "churn":
		// 200,000 abandoned 4 KiB buffers: 800 MiB allocated, ~4 KiB live.
		n = churnBuffers(200_000)
	case "retain":
		// 8,000 retained 4 KiB buffers: 32 MiB allocated, all of it live.
		n = retainBuffers(8_000)
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q: want churn or retain\n", *mode)
		os.Exit(2)
	}
	elapsed := time.Since(start)

	// The collector has to run before the profile is written, or inuse_space
	// reports objects that are unreachable but not yet swept — which looks
	// exactly like a leak.
	runtime.GC()

	f, err := os.Create(*out)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := pprof.WriteHeapProfile(f); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := f.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	fmt.Printf("mode            %s\n", *mode)
	fmt.Printf("MemProfileRate  %d bytes\n", runtime.MemProfileRate)
	fmt.Printf("elapsed         %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("work result     %d\n", n)
	fmt.Printf("alloc_space     %s over %d objects\n", mib(ms.TotalAlloc), ms.Mallocs)
	fmt.Printf("inuse_space     %s over %d objects\n", mib(ms.HeapAlloc), ms.HeapObjects)
	fmt.Printf("profile         %s\n", *out)
}

func mib(b uint64) string { return fmt.Sprintf("%.1f MiB", float64(b)/(1<<20)) }
