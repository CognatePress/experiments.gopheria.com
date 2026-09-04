// Package mutexchan measures one shared counter incremented by N goroutines at
// once, guarded five different ways, so that "share memory by communicating"
// can be read as a curve instead of as a slogan.
//
// The implementations are deliberately equivalent in what they guarantee, with
// one exception that is labelled where it appears: an atomic add serialises the
// increment but cannot contain a critical section, so it is measured only with
// an empty one and stands as a floor rather than as a competitor.
package mutexchan

import (
	"sync"
	"sync/atomic"
)

// Counter is a shared integer that many goroutines increment concurrently.
//
// Handle returns the closure a single goroutine will call. An implementation
// that needs per-goroutine state — a reply channel, say — allocates it here, so
// that cost is paid once per goroutine rather than once per operation.
//
// Close stops any owner goroutine and returns the final value. The benchmark
// checks it against the operation count, which is what proves an asynchronous
// implementation finished the work rather than dropping it on the floor.
type Counter interface {
	Handle() func()
	Close() int64
}

// sink keeps the result of the in-section work reachable. Without it the
// compiler is free to delete the work and the critical section measures the
// lock alone.
var sink atomic.Uint64

// spin is the work performed while the counter is held. xorshift64 was chosen
// because it is compute-bound, has no memory traffic to confuse the cache
// story, and costs a predictable amount per iteration.
//
//go:noinline
func spin(n int) uint64 {
	x := uint64(0x9e3779b97f4a7c15)
	for i := 0; i < n; i++ {
		x ^= x << 13
		x ^= x >> 7
		x ^= x << 17
	}
	return x
}

// mutexCounter is the shared-memory version: every caller takes the lock and
// does the work itself.
type mutexCounter struct {
	mu   sync.Mutex
	v    int64
	acc  uint64
	work int
}

// NewMutex returns a counter guarded by a sync.Mutex, doing work iterations of
// spin inside the critical section.
func NewMutex(work int) Counter { return &mutexCounter{work: work} }

func (c *mutexCounter) Handle() func() {
	return func() {
		c.mu.Lock()
		c.v++
		if c.work > 0 {
			c.acc += spin(c.work)
		}
		c.mu.Unlock()
	}
}

func (c *mutexCounter) Close() int64 {
	sink.Store(c.acc)
	return c.v
}

// atomicCounter is the floor: no lock, no owner, no critical section.
type atomicCounter struct{ v atomic.Int64 }

// NewAtomic returns a counter that is a single atomic add. It admits no work
// inside the increment, so the benchmark measures it only with an empty
// critical section.
func NewAtomic() Counter { return &atomicCounter{} }

func (c *atomicCounter) Handle() func() { return func() { c.v.Add(1) } }

func (c *atomicCounter) Close() int64 { return c.v.Load() }

// chanCounter is the communicating version: one goroutine owns the value and
// every other goroutine sends it an increment. Nothing waits for the increment
// to be applied, so a send returns as soon as the owner — or the buffer — has
// taken it.
type chanCounter struct {
	ops  chan struct{}
	done chan struct{}
	v    int64
	acc  uint64
	work int
}

// NewChan returns a channel-owned counter with a buffer of buf. buf=0 is the
// synchronous handoff; a large buf decouples the sender from the owner until
// the owner falls behind.
func NewChan(buf, work int) Counter {
	c := &chanCounter{
		ops:  make(chan struct{}, buf),
		done: make(chan struct{}),
		work: work,
	}
	go func() {
		for range c.ops {
			c.v++
			if c.work > 0 {
				c.acc += spin(c.work)
			}
		}
		close(c.done)
	}()
	return c
}

func (c *chanCounter) Handle() func() { return func() { c.ops <- struct{}{} } }

// Close drains the channel before reporting, so a buffered counter cannot look
// fast by leaving work undone.
func (c *chanCounter) Close() int64 {
	close(c.ops)
	<-c.done
	sink.Store(c.acc)
	return c.v
}

// ackCounter is the communicating version with the mutex's guarantee: the
// caller waits until the owner has applied the increment. That is what makes it
// comparable to a Lock/Unlock pair rather than to a fire-and-forget send.
type ackCounter struct {
	reqs chan chan struct{}
	done chan struct{}
	v    int64
	acc  uint64
	work int
}

// NewAck returns a channel-owned counter whose callers wait for the increment
// to be applied.
func NewAck(work int) Counter {
	c := &ackCounter{
		reqs: make(chan chan struct{}),
		done: make(chan struct{}),
		work: work,
	}
	go func() {
		for reply := range c.reqs {
			c.v++
			if c.work > 0 {
				c.acc += spin(c.work)
			}
			reply <- struct{}{}
		}
		close(c.done)
	}()
	return c
}

// Handle allocates one reply channel per goroutine rather than one per call.
// Allocating per call is the version everyone writes first, and it would put an
// allocation in every row of the table instead of the contention this is
// measuring.
func (c *ackCounter) Handle() func() {
	reply := make(chan struct{})
	return func() {
		c.reqs <- reply
		<-reply
	}
}

func (c *ackCounter) Close() int64 {
	close(c.reqs)
	<-c.done
	sink.Store(c.acc)
	return c.v
}
