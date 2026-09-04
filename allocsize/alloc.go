// Package allocsize prices one heap allocation at each of the size classes
// either side of 80 bytes.
//
// Go 1.27 makes the compiler emit size-specialised allocation routines and
// claims up to 30% off allocations under 80 bytes. The specialisation is
// chosen at compile time from a constant size, so every benchmark here
// allocates a literal size — a size read from a slice would defeat the thing
// being measured and the table would come back flat for the wrong reason.
package allocsize

import "unsafe"

// keep stores the address somewhere the compiler cannot prove dead, which is
// what puts the object on the heap. Without it escape analysis keeps the small
// sizes in the frame and the benchmark measures nothing — the failure L2 was
// written about.
//
//go:noinline
func keep(p unsafe.Pointer) { sinkPtr = p }

var (
	sinkPtr unsafe.Pointer
	sinkB   []byte
)
