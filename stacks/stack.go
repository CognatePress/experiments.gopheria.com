// Package stacks probes how a goroutine's stack grows: where the runtime copies
// it, how large each new one is, and what happens to a pointer that was aimed
// at the old one.
package stacks

import (
	"runtime"
	"unsafe"
)

// FrameBytes is the size of the local array each recursion level puts on the
// stack. It is what turns a recursion depth into a byte count.
const FrameBytes = 128

var sinkByte byte

// Descend recurses len(addrs) levels deep and records the address of each
// frame's local array. Consecutive frames sit a fixed stride apart; a stride
// that is not that constant is the runtime having moved the stack.
//
//go:noinline
func Descend(depth int, addrs []uintptr) {
	var frame [FrameBytes]byte
	addrs[depth] = uintptr(unsafe.Pointer(&frame))
	frame[0] = byte(depth)
	if depth+1 < len(addrs) {
		Descend(depth+1, addrs)
	}
	sinkByte = frame[0]
}

// Move is a copy event: the depth at which the frame address jumped, the stride
// between two frames, and the bytes of frames live at that moment. A copy moves
// every frame, so the live total is every frame so far — not only the ones
// pushed since the last copy.
type Move struct {
	Depth  int
	Stride int
	Used   int
}

// Moves reduces a run of frame addresses to the copy events inside it. The
// stride between two frames within one segment is constant, so a delta that is
// not the stride means the frames now live somewhere else.
func Moves(addrs []uintptr) []Move {
	if len(addrs) < 3 {
		return nil
	}
	stride := diff(addrs[0], addrs[1])
	var moves []Move
	for i := 1; i < len(addrs); i++ {
		if diff(addrs[i-1], addrs[i]) == stride {
			continue
		}
		moves = append(moves, Move{Depth: i, Stride: stride, Used: i * stride})
	}
	return moves
}

func diff(a, b uintptr) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}

// StackInuseAt parks a goroutine at the given recursion depth and reports the
// process-wide stack bytes while it is parked, minus the reading taken before
// it started. ReadMemStats stops the world, so this is read from outside the
// recursing goroutine rather than from inside it.
func StackInuseAt(depth int) uint64 {
	runtime.GC()
	var before, during runtime.MemStats
	runtime.ReadMemStats(&before)

	parked := make(chan struct{})
	release := make(chan struct{})
	go func() {
		addrs := make([]uintptr, depth)
		descendThenPark(0, addrs, parked, release)
	}()
	<-parked
	runtime.ReadMemStats(&during)
	close(release)
	return during.StackInuse - before.StackInuse
}

//go:noinline
func descendThenPark(depth int, addrs []uintptr, parked, release chan struct{}) {
	var frame [FrameBytes]byte
	addrs[depth] = uintptr(unsafe.Pointer(&frame))
	frame[0] = byte(depth)
	if depth+1 < len(addrs) {
		descendThenPark(depth+1, addrs, parked, release)
	} else {
		close(parked)
		<-release
	}
	sinkByte = frame[0]
}
