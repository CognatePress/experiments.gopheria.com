package main

import (
	"sync/atomic"
	"testing"
)

var sink atomic.Uint64

// BenchmarkMix is run across a -cpu list so that each row is one value of
// GOMAXPROCS against the same fixed unit of work.
//
//	go test -run=^$ -bench=Mix -benchmem -count=10 \
//	  -cpu=1,2,3,4,6,8,10,12,14,16,20,24 ./gomaxprocs
func BenchmarkMix(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var acc uint64
		for pb.Next() {
			acc ^= mix(acc+1, 512)
		}
		sink.Add(acc)
	})
}
