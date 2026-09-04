package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"
)

// syscallLeaf is the case correct plumbing does not fix. The context reaches
// the leaf, the cancellation fires on time, and the goroutine stays blocked
// anyway — because it is parked inside a read(2) on a pipe, which is a
// descriptor the netpoller does not own.
//
// This is L1's finding arriving from the application side: a goroutine blocked
// in a syscall the runtime cannot park costs an OS thread, and cancellation is
// cooperative, so nothing about ctx.Done() reaches it.
func syscallLeaf(wait time.Duration) {
	before := runtime.NumGoroutine()

	r, w, err := os.Pipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	blocked := make(chan struct{})
	returned := make(chan error, 1)

	go func() {
		// A raw file descriptor, not an *os.File the runtime registered with
		// the netpoller: os.NewFile on a pipe gives back a pollable descriptor
		// on darwin, so the syscall is issued through the raw fd instead.
		fd := int(r.Fd())
		buf := make([]byte, 1)
		close(blocked)
		returned <- rawRead(fd, buf)
	}()

	<-blocked
	time.Sleep(50 * time.Millisecond)
	during := runtime.NumGoroutine()

	cancelAt := time.Now()
	cancel()
	<-ctx.Done()
	ctxLatency := time.Since(cancelAt)

	select {
	case err := <-returned:
		fmt.Printf("reader returned after cancel: %v\n", err)
	case <-time.After(wait):
		fmt.Printf("reader still blocked %s after cancel\n", wait)
	}

	after := runtime.NumGoroutine()

	fmt.Printf("mode              syscall\n")
	fmt.Printf("goroutines        %d before, %d with reader blocked, %d after cancel\n", before, during, after)
	fmt.Printf("ctx.Done latency  %s\n", ctxLatency.Round(time.Microsecond))
	fmt.Printf("leaked            %d\n", after-before)

	// Closing the write end is the only thing that releases the reader. That is
	// the actual remedy: cancellation has to be wired to whatever the leaf is
	// blocked on, not only to the context.
	closeAt := time.Now()
	if err := w.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	select {
	case err := <-returned:
		fmt.Printf("reader released   %s after closing the write end (err=%v)\n",
			time.Since(closeAt).Round(time.Microsecond), err)
	case <-time.After(wait):
		fmt.Printf("reader still blocked after closing the write end\n")
	}
	if err := r.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
