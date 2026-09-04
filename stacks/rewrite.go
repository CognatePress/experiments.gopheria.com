package stacks

import "unsafe"

// Rewritten is what one stack copy did to a pointer and to an integer holding
// the same address at the moment the copy happened.
type Rewritten struct {
	Before  uintptr // the address before the stack grew
	Pointer uintptr // where the pointer points afterwards
	Stale   uintptr // the uintptr that was copied out of it beforehand
	Value   int     // what the pointer reads afterwards
	Moves   int     // stack copies the descent triggered
}

// Rewrite takes the address of a local, forces the stack under it to grow, and
// reports where the pointer ended up. The runtime rewrites every pointer that
// aimed into the old stack; a uintptr is not a pointer and is left alone.
//
//go:noinline
func Rewrite(depth int) Rewritten {
	anchor := 42
	p := &anchor

	r := Rewritten{Before: uintptr(unsafe.Pointer(p))}
	stale := r.Before // a plain integer from here on

	addrs := make([]uintptr, depth)
	Descend(0, addrs)

	r.Pointer = uintptr(unsafe.Pointer(p))
	r.Stale = stale
	r.Value = *p
	r.Moves = len(Moves(addrs))
	return r
}
