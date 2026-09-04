// Command errgrouplab prints what errgroup actually does when more than one
// sibling fails.
//
// It looks like a WaitGroup that collects errors and behaves like neither:
// Wait returns the first error rather than all of them, WithContext cancels the
// siblings on that first error, and SetLimit blocks the caller of Go instead of
// queueing — which is where the deadlocks come from.
//
//	go build -o /tmp/errgrouplab ./errgrouplab
//	/tmp/errgrouplab -mode=first
//	/tmp/errgrouplab -mode=nocontext
//	/tmp/errgrouplab -mode=limit
//	/tmp/errgrouplab -mode=deadlock
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

// start is the process clock every printed timestamp is relative to, so the
// ordering of events is readable without subtracting wall clocks by hand.
var start = time.Now()

func at(format string, args ...any) {
	fmt.Printf("%7s  %s\n", time.Since(start).Round(time.Millisecond), fmt.Sprintf(format, args...))
}

func main() {
	mode := flag.String("mode", "first", "first, nocontext, limit or deadlock")
	flag.Parse()

	switch *mode {
	case "first":
		first()
	case "nocontext":
		nocontext()
	case "limit":
		limit()
	case "deadlock":
		deadlock()
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q\n", *mode)
		os.Exit(2)
	}
}

// sibling fails after d, unless its context is cancelled first. It reports
// which of the two happened, which is what makes the cancellation visible.
func sibling(ctx context.Context, name string, d time.Duration, observed *atomic.Int64) func() error {
	return func() error {
		select {
		case <-time.After(d):
			at("%s fails", name)
			observed.Add(1)
			return fmt.Errorf("%s failed", name)
		case <-ctx.Done():
			at("%s cancelled: %v", name, ctx.Err())
			return ctx.Err()
		}
	}
}

// first: three siblings fail at 50ms, 100ms and 150ms. Only one of those
// failures is ever reached.
func first() {
	g, ctx := errgroup.WithContext(context.Background())
	var reached atomic.Int64

	g.Go(sibling(ctx, "alpha", 50*time.Millisecond, &reached))
	g.Go(sibling(ctx, "beta", 100*time.Millisecond, &reached))
	g.Go(sibling(ctx, "gamma", 150*time.Millisecond, &reached))

	// A watcher on the derived context, so the moment it closes is printed
	// rather than inferred from the sibling that noticed.
	go func() {
		<-ctx.Done()
		at("derived context cancelled: %v", ctx.Err())
	}()

	err := g.Wait()
	at("Wait returned: %v", err)
	at("errors.Is(err, context.Canceled) = %v", errors.Is(err, context.Canceled))
	at("siblings that reached their own failure: %d of 3", reached.Load())
}

// nocontext: the same three siblings under a bare errgroup.Group. Nothing
// cancels anything, so every sibling runs to its own failure and the two later
// errors are discarded in silence.
func nocontext() {
	var g errgroup.Group
	var reached atomic.Int64
	ctx := context.Background()

	g.Go(sibling(ctx, "alpha", 50*time.Millisecond, &reached))
	g.Go(sibling(ctx, "beta", 100*time.Millisecond, &reached))
	g.Go(sibling(ctx, "gamma", 150*time.Millisecond, &reached))

	err := g.Wait()
	at("Wait returned: %v", err)
	at("siblings that reached their own failure: %d of 3", reached.Load())
}

// limit: SetLimit(2) with three tasks. The third call to Go does not queue the
// task — it blocks the goroutine that called it until a slot frees.
func limit() {
	var g errgroup.Group
	g.SetLimit(2)

	for i, d := range []time.Duration{80 * time.Millisecond, 80 * time.Millisecond, 10 * time.Millisecond} {
		name := fmt.Sprintf("task%d", i+1)
		at("calling Go for %s", name)
		g.Go(func() error {
			at("%s running", name)
			time.Sleep(d)
			at("%s done", name)
			return nil
		})
		at("Go for %s returned", name)
	}

	if err := g.Wait(); err != nil {
		at("Wait returned: %v", err)
		return
	}
	at("Wait returned: <nil>")
}

// deadlock: the failure mode SetLimit's blocking behaviour produces. A task
// holding one of the two slots waits for a result that can only be produced by
// a task that cannot start, because starting it needs the slot the waiter is
// holding.
//
// TryGo is the escape hatch: it declines rather than blocks. The program uses
// it to show the deadlock without becoming one.
func deadlock() {
	var g errgroup.Group
	g.SetLimit(1)

	g.Go(func() error {
		at("outer running, holding the only slot")
		// This is the call that would deadlock. TryGo reports the refusal
		// instead of waiting for a slot that the caller itself is holding.
		if g.TryGo(func() error { return nil }) {
			at("inner started — no limit reached")
			return nil
		}
		at("TryGo refused: no slot free, and Go here would block forever")
		return errors.New("nested Go under SetLimit(1)")
	})

	err := g.Wait()
	at("Wait returned: %v", err)
}
