// Command gclab runs one steady allocator and reports what the collector did
// while it ran.
//
// It is the workload behind three articles: the GOGC/GOMEMLIMIT comparison, the
// Green Tea A/B, and the gctrace annotation. All three need the same thing — a
// program whose allocation rate, live set and pointer density are fixed from
// the command line, so that the only variable is the knob under test.
//
//	go build -o /tmp/gclab ./gclab
//	GOGC=100 /tmp/gclab -mode=pointer -live=256MB -dur=10s
//	GOGC=off GOMEMLIMIT=512MiB /tmp/gclab -mode=pointer -live=256MB -dur=10s
//	GODEBUG=gctrace=1 /tmp/gclab -mode=pointer -live=256MB -dur=10s
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/metrics"
	"time"
)

func main() {
	mode := flag.String("mode", "pointer", "object shape: pointer, scalar or graph")
	live := flag.Int64("live", 256<<20, "bytes retained in the live set")
	dur := flag.Duration("dur", 10*time.Second, "how long to churn")
	label := flag.String("label", "", "label reproduced in the report line")
	flag.Parse()

	switch *mode {
	case "pointer", "scalar", "graph":
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q: want pointer, scalar or graph\n", *mode)
		os.Exit(2)
	}

	// The sampler runs alongside the workload because the two numbers that
	// matter most are peaks, and a reading taken after the loop has finished
	// reports whatever the last collection left behind instead.
	stop := make(chan struct{})
	peaks := make(chan peak)
	go sample(stop, peaks)

	before := read()
	start := time.Now()
	ops := churn(*mode, *live, *dur)
	elapsed := time.Since(start)
	after := read()

	close(stop)
	pk := <-peaks

	report(*label, *mode, ops, elapsed, before, after, pk)
}

// The metrics read here are the ones that answer "what did the collector do",
// as opposed to "how big is the heap right now":
//
//   - gc/cycles/total     how many collections ran
//   - gc/heap/goal        the target the pacer is aiming at, which is what GOGC
//     sets and what GOMEMLIMIT overrides
//   - cpu/classes/gc      CPU seconds spent collecting, against the total, which
//     is the only honest way to say "the GC cost 8%"
//   - memory/classes/total the process's mapped total, the closest thing the
//     runtime reports to RSS
var samples = []string{
	"/gc/cycles/total:gc-cycles",
	"/gc/heap/goal:bytes",
	"/gc/heap/live:bytes",
	"/cpu/classes/gc/total:cpu-seconds",
	"/cpu/classes/total:cpu-seconds",
	"/cpu/classes/gc/mark/assist:cpu-seconds",
	"/cpu/classes/gc/mark/dedicated:cpu-seconds",
	"/cpu/classes/gc/mark/idle:cpu-seconds",
	"/cpu/classes/gc/pause:cpu-seconds",
	"/memory/classes/total:bytes",
	"/memory/classes/heap/free:bytes",
	"/memory/classes/heap/released:bytes",
}

// The four /cpu/classes/gc children are read separately because only three of
// them are governed by the runtime's GC CPU limiter. Mark work done on an
// otherwise idle P is charged to the GC in /cpu/classes/gc/total but costs the
// application nothing it was using, so a share computed from that total can sit
// well above the limiter's own ceiling without the limiter being wrong.

type reading struct {
	cycles      uint64
	heapGoal    uint64
	heapLive    uint64
	gcCPU       float64
	totalCPU    float64
	mapped      uint64
	heapFree    uint64
	heapFreed   uint64
	assistCPU   float64
	dedicCPU    float64
	idleMarkCPU float64
	pauseCPU    float64
	pauseTotal  time.Duration
}

func read() reading {
	ms := make([]metrics.Sample, len(samples))
	for i, name := range samples {
		ms[i].Name = name
	}
	metrics.Read(ms)

	// By name, not by position. An earlier version indexed this slice by
	// number, and inserting one metric in the middle re-pointed every field
	// after it at the wrong sample — silently for the byte counts, and with a
	// panic for the first float read as a uint64.
	u := func(name string) uint64 {
		for i := range ms {
			if ms[i].Name == name {
				return ms[i].Value.Uint64()
			}
		}
		panic("no such metric: " + name)
	}
	f := func(name string) float64 {
		for i := range ms {
			if ms[i].Name == name {
				return ms[i].Value.Float64()
			}
		}
		panic("no such metric: " + name)
	}

	var ms2 runtime.MemStats
	runtime.ReadMemStats(&ms2)

	return reading{
		cycles:      u("/gc/cycles/total:gc-cycles"),
		heapGoal:    u("/gc/heap/goal:bytes"),
		heapLive:    u("/gc/heap/live:bytes"),
		gcCPU:       f("/cpu/classes/gc/total:cpu-seconds"),
		totalCPU:    f("/cpu/classes/total:cpu-seconds"),
		mapped:      u("/memory/classes/total:bytes"),
		heapFree:    u("/memory/classes/heap/free:bytes"),
		heapFreed:   u("/memory/classes/heap/released:bytes"),
		assistCPU:   f("/cpu/classes/gc/mark/assist:cpu-seconds"),
		dedicCPU:    f("/cpu/classes/gc/mark/dedicated:cpu-seconds"),
		idleMarkCPU: f("/cpu/classes/gc/mark/idle:cpu-seconds"),
		pauseCPU:    f("/cpu/classes/gc/pause:cpu-seconds"),
		pauseTotal:  time.Duration(ms2.PauseTotalNs),
	}
}

type peak struct {
	goal   uint64
	mapped uint64
	live   uint64
}

// sample polls at 10ms. The pacer moves the heap goal on every cycle, and at
// the allocation rates this workload reaches a 100ms poll misses most of them.
func sample(stop <-chan struct{}, out chan<- peak) {
	t := time.NewTicker(10 * time.Millisecond)
	defer t.Stop()
	var p peak
	for {
		select {
		case <-stop:
			out <- p
			return
		case <-t.C:
			r := read()
			p.goal = max(p.goal, r.heapGoal)
			p.mapped = max(p.mapped, r.mapped)
			p.live = max(p.live, r.heapLive)
		}
	}
}

func report(label, mode string, ops int64, elapsed time.Duration, before, after reading, pk peak) {
	cycles := after.cycles - before.cycles
	gcCPU := after.gcCPU - before.gcCPU
	totalCPU := after.totalCPU - before.totalCPU

	assist := after.assistCPU - before.assistCPU
	dedicated := after.dedicCPU - before.dedicCPU
	idleMark := after.idleMarkCPU - before.idleMarkCPU
	pauseCPU := after.pauseCPU - before.pauseCPU

	var share, limited float64
	if totalCPU > 0 {
		share = gcCPU / totalCPU * 100
		// What the runtime's ~50% limiter actually governs: everything the GC
		// took from the application. Idle-time marking is excluded because it
		// ran on CPU the application was not using.
		limited = (gcCPU - idleMark) / totalCPU * 100
	}
	var perCycle time.Duration
	if cycles > 0 {
		perCycle = elapsed / time.Duration(cycles)
	}

	if label != "" {
		fmt.Printf("label        %s\n", label)
	}
	fmt.Printf("mode         %s\n", mode)
	fmt.Printf("GOGC         %s\n", envOr("GOGC", "100 (default)"))
	fmt.Printf("GOMEMLIMIT   %s\n", envOr("GOMEMLIMIT", "math.MaxInt64 (default)"))
	fmt.Printf("GOEXPERIMENT %s\n", envOr("GOEXPERIMENT", "(none)"))
	fmt.Printf("elapsed      %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("allocations  %d (%.1f M/s)\n", ops, float64(ops)/elapsed.Seconds()/1e6)
	fmt.Printf("gc cycles    %d (one every %s)\n", cycles, perCycle.Round(time.Microsecond))
	fmt.Printf("gc cpu       %.2fs of %.2fs = %.1f%%\n", gcCPU, totalCPU, share)
	fmt.Printf("gc cpu net   %.1f%% excluding idle-time marking (what the ~50%% limiter governs)\n", limited)
	fmt.Printf("gc cpu split assist %.2fs · dedicated %.2fs · idle %.2fs · pause %.2fs\n",
		assist, dedicated, idleMark, pauseCPU)
	fmt.Printf("stw total    %s\n", (after.pauseTotal - before.pauseTotal).Round(time.Microsecond))
	fmt.Printf("heap goal    peak %s, final %s\n", mib(pk.goal), mib(after.heapGoal))
	fmt.Printf("heap live    peak %s, final %s\n", mib(pk.live), mib(after.heapLive))
	fmt.Printf("mapped       peak %s, final %s\n", mib(pk.mapped), mib(after.mapped))
}

func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func mib(b uint64) string { return fmt.Sprintf("%.1f MiB", float64(b)/(1<<20)) }
