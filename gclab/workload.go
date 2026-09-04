package main

import (
	"math/rand/v2"
	"time"
)

// The workload is a steady allocator: a live set of fixed size, and a churn
// loop that replaces one member of it per iteration. The live set never grows,
// so the heap reaches a plateau and stays there — which is the shape that makes
// GOGC and GOMEMLIMIT comparable, because a heap that is still growing hides
// which knob set the ceiling.

// The two object shapes are the same size and cost the same number of
// allocations. They differ only in what the mark phase finds inside them:
// ptrObj is 96 pointer slots the collector must load and follow, scalarObj is
// 768 bytes it may skip entirely.
//
// Getting this right matters more than it looks. An earlier version built the
// pointer object as a chain of sixteen linked nodes, which made it sixteen
// allocations against the scalar object's one — so the two modes differed in
// allocation rate as well as in pointer density, and no delta could be
// attributed to either.
const (
	objBytes = 768
	ptrSlots = objBytes / 8 // 96
)

type ptrObj struct{ p [ptrSlots]*byte }

type scalarObj struct{ b [objBytes]byte }

// pointees is what every ptrObj points into. One shared array rather than
// per-object targets, so the two modes hold the same number of live bytes:
// what is being varied is the scanning work, not the size of the heap.
var pointees = make([]byte, ptrSlots)

// object is what the live set holds. Exactly one of the two fields is set.
type object struct {
	p *ptrObj
	s *scalarObj
}

func newPtr(seed byte) object {
	o := new(ptrObj)
	for i := range o.p {
		o.p[i] = &pointees[i]
	}
	*o.p[0] = seed
	return object{p: o}
}

func newScalar(seed byte) object {
	o := new(scalarObj)
	o.b[0] = seed
	return object{s: o}
}

// churn holds liveBytes worth of objects and replaces one per iteration until
// dur elapses. It returns the number of replacements, which is the allocation
// count every other number in the report is divided by.
func churn(mode string, liveBytes int64, dur time.Duration) int64 {
	alloc := newPtr
	if mode == "scalar" {
		alloc = newScalar
	}

	n := int(liveBytes / objBytes)
	if n < 1 {
		n = 1
	}
	live := make([]object, n)
	for i := range live {
		live[i] = alloc(byte(i))
	}

	// A random index rather than a sweep, so the collector sees objects of
	// mixed age at every cycle instead of a wavefront moving through the slice.
	rng := rand.New(rand.NewPCG(0x9e3779b9, 0x7f4a7c15))

	var ops int64
	deadline := time.Now().Add(dur)
	for {
		// The clock is read once per batch. time.Now costs about as much as the
		// allocation being timed, and checking it every iteration would make
		// this a benchmark of time.Now.
		for i := 0; i < 512; i++ {
			live[rng.IntN(n)] = alloc(byte(ops))
			ops++
		}
		if time.Now().After(deadline) {
			return ops
		}
	}
}
