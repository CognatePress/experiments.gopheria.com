// Command fairness answers a question the sweep cannot: does sync.Mutex ever
// enter starvation mode under this workload?
//
// The mode is not observable from outside the runtime, so this measures its two
// symptoms instead. In normal mode an arriving goroutine may barge ahead of
// queued waiters, which makes the per-goroutine operation counts uneven. In
// starvation mode the lock is handed off FIFO and the counts even out. The
// switch happens when a waiter has been queued for more than
// starvationThresholdNs, which is 1ms — so the second measurement is simply the
// longest Lock any goroutine waited for.
//
//	go build -o /tmp/fairness ./mutexchan/fairness
//	/tmp/fairness -dur=2s
//	/tmp/fairness -dur=2s -timed
package main

import (
	"flag"
	"fmt"
	"sort"
	"sync"
	"time"
)

var counts = []int{2, 4, 8, 16, 32, 64}

func main() {
	dur := flag.Duration("dur", 2*time.Second, "how long each contention level runs")
	timed := flag.Bool("timed", false, "time every Lock and report the longest wait")
	flag.Parse()

	if *timed {
		fmt.Printf("%-4s  %12s  %12s  %12s   %s\n", "n", "max wait", "p99 wait", "median", "reached 1ms")
	} else {
		fmt.Printf("%-4s  %12s  %12s  %12s  %8s\n", "n", "min ops", "median ops", "max ops", "max/min")
	}

	for _, n := range counts {
		if *timed {
			max, p99, med := timedRun(n, *dur)
			fmt.Printf("%-4d  %12s  %12s  %12s   %v\n",
				n, max.Round(time.Microsecond), p99.Round(time.Microsecond),
				med.Round(time.Nanosecond), max >= time.Millisecond)
			continue
		}
		lo, med, hi := countRun(n, *dur)
		fmt.Printf("%-4d  %12d  %12d  %12d  %8.2f\n", n, lo, med, hi, float64(hi)/float64(lo))
	}
}

// countRun measures how evenly the lock was distributed. Nothing is timed, so
// the loop is the same one the sweep benchmarks.
func countRun(n int, dur time.Duration) (lo, med, hi int64) {
	var mu sync.Mutex
	var shared int64

	ops := make([]int64, n)
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			var local int64
			for {
				select {
				case <-stop:
					ops[i] = local
					return
				default:
				}
				// A batch between stop checks: reading the channel every
				// iteration would cost more than the lock being measured.
				for j := 0; j < 1024; j++ {
					mu.Lock()
					shared++
					mu.Unlock()
					local++
				}
			}
		}(i)
	}

	time.Sleep(dur)
	close(stop)
	wg.Wait()

	sorted := append([]int64(nil), ops...)
	sort.Slice(sorted, func(a, b int) bool { return sorted[a] < sorted[b] })
	return sorted[0], sorted[len(sorted)/2], sorted[len(sorted)-1]
}

// timedRun measures the wait itself. Two time.Now calls per Lock cost more than
// the Lock does, so the absolute numbers are inflated; the question this answers
// is only whether any waiter crosses the 1ms threshold, and an inflated
// measurement that stays below it is the stronger answer.
func timedRun(n int, dur time.Duration) (max, p99, med time.Duration) {
	var mu sync.Mutex
	var shared int64

	waits := make([][]time.Duration, n)
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			local := make([]time.Duration, 0, 1<<16)
			for {
				select {
				case <-stop:
					waits[i] = local
					return
				default:
				}
				for j := 0; j < 1024; j++ {
					start := time.Now()
					mu.Lock()
					local = append(local, time.Since(start))
					shared++
					mu.Unlock()
				}
			}
		}(i)
	}

	time.Sleep(dur)
	close(stop)
	wg.Wait()

	var all []time.Duration
	for _, w := range waits {
		all = append(all, w...)
	}
	sort.Slice(all, func(a, b int) bool { return all[a] < all[b] })
	if len(all) == 0 {
		return 0, 0, 0
	}
	return all[len(all)-1], all[len(all)*99/100], all[len(all)/2]
}
