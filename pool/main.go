// Command pool measures how many objects survive a sync.Pool across GC cycles.
//
// New() is counted, so "misses" is exactly the number of Gets the pool could
// not satisfy from what was Put into it.
package main

import (
	"fmt"
	"runtime"
	"sync"
)

const objects = 100

type buf struct{ b [1024]byte }

var misses int

func newPool() *sync.Pool {
	misses = 0
	return &sync.Pool{New: func() any { misses++; return new(buf) }}
}

// fill puts `objects` items in, then reports how many Gets had to call New.
func drain(p *sync.Pool) int {
	misses = 0
	for i := 0; i < objects; i++ {
		p.Get()
	}
	return misses
}

func fill(p *sync.Pool) {
	for i := 0; i < objects; i++ {
		p.Put(new(buf))
	}
}

func main() {
	// One P, one goroutine: sync.Pool is per-P, so this keeps the result
	// deterministic instead of depending on where the goroutine lands.
	runtime.GOMAXPROCS(1)

	p := newPool()

	fill(p)
	fmt.Printf("no GC:      %3d/%d Gets missed\n", drain(p), objects)

	fill(p)
	runtime.GC()
	fmt.Printf("after 1 GC: %3d/%d Gets missed\n", drain(p), objects)

	fill(p)
	runtime.GC()
	runtime.GC()
	fmt.Printf("after 2 GC: %3d/%d Gets missed\n", drain(p), objects)
}
