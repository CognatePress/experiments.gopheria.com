// Command ctxleak measures what happens to a three-layer call chain when the
// middle layer stops passing the context down.
//
// The plumbing is the part everyone gets right in review and wrong in a hurry:
// a function accepts a ctx, starts a goroutine, and hands that goroutine a
// fresh Background instead of the one it was given. Everything below it is then
// uncancellable, and nothing about the signature says so.
//
//	go build -o /tmp/ctxleak ./ctxleak
//	/tmp/ctxleak -mode=propagate
//	/tmp/ctxleak -mode=dropped
//	/tmp/ctxleak -mode=syscall
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"
)

// leaves is how many goroutines the bottom layer starts. A leak is easier to
// read as a count than as a single stuck goroutine.
const leaves = 200

// leaf blocks until its context is cancelled, then reports how long that took.
func leaf(ctx context.Context, started time.Time, done chan<- time.Duration) {
	<-ctx.Done()
	done <- time.Since(started)
}

// middle is the layer under test. When drop is true it does the thing this
// article is about: it accepts a context and passes a different one on.
func middle(ctx context.Context, drop bool, started time.Time, done chan<- time.Duration) {
	down := ctx
	if drop {
		down = context.Background()
	}
	for i := 0; i < leaves; i++ {
		go leaf(down, started, done)
	}
}

func main() {
	mode := flag.String("mode", "propagate", "propagate, dropped or syscall")
	wait := flag.Duration("wait", 2*time.Second, "how long to wait for the leaves after cancelling")
	flag.Parse()

	switch *mode {
	case "propagate", "dropped":
		chain(*mode == "dropped", *wait)
	case "syscall":
		syscallLeaf(*wait)
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q\n", *mode)
		os.Exit(2)
	}
}

func chain(drop bool, wait time.Duration) {
	before := runtime.NumGoroutine()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan time.Duration, leaves)

	started := time.Now()
	middle(ctx, drop, started, done)

	// Give the leaves time to reach their <-ctx.Done() before cancelling, so
	// the latency measured is the cancellation and not the goroutine start.
	time.Sleep(50 * time.Millisecond)
	during := runtime.NumGoroutine()

	cancelAt := time.Now()
	cancel()

	var returned int
	var slowest time.Duration
	deadline := time.After(wait)
collect:
	for returned < leaves {
		select {
		case d := <-done:
			returned++
			// d is measured from started; the cancellation latency is what is
			// left after the sleep above.
			if lat := d - cancelAt.Sub(started); lat > slowest {
				slowest = lat
			}
		case <-deadline:
			break collect
		}
	}

	// The collector cannot free a goroutine that is parked on a channel nobody
	// will ever close, so a count taken after a GC is still the leak.
	runtime.GC()
	after := runtime.NumGoroutine()

	fmt.Printf("mode              %s\n", map[bool]string{true: "dropped", false: "propagate"}[drop])
	fmt.Printf("goroutines        %d before, %d with leaves running, %d after cancel\n", before, during, after)
	fmt.Printf("leaves returned   %d of %d\n", returned, leaves)
	fmt.Printf("slowest cancel    %s\n", slowest.Round(time.Microsecond))
	fmt.Printf("leaked            %d\n", after-before)
}
