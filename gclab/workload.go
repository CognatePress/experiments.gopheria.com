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

// gnode is the third shape, and the one Green Tea is aimed at: a small object
// carrying a couple of pointers that lead somewhere scattered. ptrObj is
// pointer-dense but its 96 slots all address one contiguous array, which is the
// easiest possible case for a mark phase. A heap of gnodes is the opposite —
// many small objects, each sending the collector somewhere it has not just
// been.
type gnode struct {
	a, b *gnode
	_    [16]byte
}

const gnodeBytes = 32

// pool holds the long-lived nodes every churned gnode points into. Keeping the
// targets alive separately is what stops the graph from retaining its own
// garbage: a fresh node points at two pool entries chosen at random, and when
// it is dropped nothing else refers to it.
var pool []*gnode

type scalarObj struct{ b [objBytes]byte }

// pointees is what every ptrObj points into. One shared array rather than
// per-object targets, so the two modes hold the same number of live bytes:
// what is being varied is the scanning work, not the size of the heap.
var pointees = make([]byte, ptrSlots)

// object is what the live set holds. Exactly one of the two fields is set.
type object struct {
	p *ptrObj
	s *scalarObj
	g *gnode
}

func newPtr(seed int64) object {
	o := new(ptrObj)
	for i := range o.p {
		o.p[i] = &pointees[i]
	}
	*o.p[0] = byte(seed)
	return object{p: o}
}

func newGraph(seed int64) object {
	n := new(gnode)
	// Two multiplicative hashes of the allocation counter. The counter is the
	// only thing varying per object, so a narrow seed here collapses the graph
	// onto a handful of targets and the mark phase gets the cache hit it was
	// supposed to miss.
	n.a = pool[uint64(seed)*0x9e3779b97f4a7c15%uint64(len(pool))]
	n.b = pool[uint64(seed)*0xc2b2ae3d27d4eb4f%uint64(len(pool))]
	return object{g: n}
}

func newScalar(seed int64) object {
	o := new(scalarObj)
	o.b[0] = byte(seed)
	return object{s: o}
}

// churn holds liveBytes worth of objects and replaces one per iteration until
// dur elapses. It returns the number of replacements, which is the allocation
// count every other number in the report is divided by.
func churn(mode string, liveBytes int64, dur time.Duration) int64 {
	alloc, each := newPtr, int64(objBytes)
	switch mode {
	case "scalar":
		alloc = newScalar
	case "graph":
		alloc, each = newGraph, gnodeBytes
		// The pool is sized independently of the live set and allocated in one
		// pass, so its members are spread across whatever spans the allocator
		// hands out rather than being adjacent by construction.
		// 2^21 nodes is 64 MiB of targets — far past any cache on this
		// machine, which is the property the mode exists for.
		pool = make([]*gnode, 1<<21)
		for i := range pool {
			pool[i] = new(gnode)
		}
		rnd := rand.New(rand.NewPCG(0x243f6a88, 0x85a308d3))
		for i := range pool {
			pool[i].a = pool[rnd.IntN(len(pool))]
			pool[i].b = pool[rnd.IntN(len(pool))]
		}
	}

	n := int(liveBytes / each)
	if n < 1 {
		n = 1
	}
	live := make([]object, n)
	for i := range live {
		live[i] = alloc(int64(i))
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
			live[rng.IntN(n)] = alloc(ops)
			ops++
		}
		if time.Now().After(deadline) {
			return ops
		}
	}
}
