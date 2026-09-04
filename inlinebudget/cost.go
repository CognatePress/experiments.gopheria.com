// Package inlinebudget prices the constructs that spend the compiler's inlining
// budget.
//
// The budget is a number in cmd/compile — 80 "cost units" as of Go 1.27 — and
// the compiler will tell you what it charged, for free, if you ask:
//
//	go build -gcflags='-m' ./inlinebudget 2>&1 | grep 'can inline'
//
// prints "can inline Base with cost 2 as: ..." for every function it accepted,
// and
//
//	go build -gcflags='-m -m' ./inlinebudget 2>&1 | grep 'cannot inline'
//
// prints the cost of the ones it rejected. Every function below is Base plus
// exactly one construct, so subtracting Base's cost from each gives that
// construct's price with nothing else in the way.
package inlinebudget

//go:noinline
func opaque(n int) int { return n }

var sink int

// Base is the reference: one parameter, one return. Every other function here
// is this plus one thing.
func Base(n int) int { return n }

// ---- one construct each, in the order a reader would guess wrong about ----

func WithAdd(n int) int { return n + 1 }

func WithCall(n int) int { return opaque(n) }

func WithIf(n int) int {
	if n > 0 {
		return n
	}
	return 0
}

func WithFor(n int) int {
	for i := 0; i < 4; i++ {
		n += i
	}
	return n
}

func WithRange(n int, xs []int) int {
	for _, x := range xs {
		n += x
	}
	return n
}

func WithSwitch(n int) int {
	switch n {
	case 0:
		return 1
	case 1:
		return 2
	}
	return n
}

func WithTypeSwitch(v any) int {
	switch t := v.(type) {
	case int:
		return t
	case string:
		return len(t)
	}
	return 0
}

func WithDefer(n int) int {
	defer func() { sink = n }()
	return n
}

func WithClosure(n int) int {
	f := func() int { return n + 1 }
	return f()
}

func WithAppend(n int, xs []int) int {
	xs = append(xs, n)
	return len(xs)
}

func WithMake(n int) int { return len(make([]int, n)) }

func WithMapIndex(n int, m map[int]int) int { return m[n] }

func WithPanic(n int) int {
	if n < 0 {
		panic("negative")
	}
	return n
}

func WithGo(n int) int {
	go func() { sink = n }()
	return n
}

func WithSelect(n int, ch chan int) int {
	select {
	case v := <-ch:
		return v
	default:
		return n
	}
}

// ---- what the budget decision actually costs ----
//
// The Grow ladder prices the budget and says nothing about speed: its body is a
// run of constant additions, which the compiler folds, so Grow26 and Grow27
// are the same two instructions with and without a call around them.
//
// This pair is the case where the decision has a consequence. Both constructors
// return a pointer to a local; both callers let that pointer die. When the
// constructor is inlined the compiler can see the value dies and keeps it in
// the frame. When it is over budget the caller cannot see inside it, so the
// pointer has to be assumed to escape and the object goes to the heap.

// small is a value big enough to be worth allocating and small enough that the
// compiler will consider keeping it in a frame.
type small struct {
	v   int
	pad [3]int
}

// makeSmall is well inside the budget.
func makeSmall(v int) *small { return &small{v: v} }

// makeSmallFat is the same constructor with enough statements bolted on to push
// it past 80. The statements are not folded away — each reads the previous
// result out of the struct — so the body is genuinely larger, which is the
// realistic version of going over budget.
func makeSmallFat(v int) *small {
	s := &small{v: v}
	s.pad[0] = s.v + 1
	s.pad[1] = s.pad[0] * 3
	s.pad[2] = s.pad[1] ^ s.v
	s.v = s.pad[2] - s.pad[0]
	s.pad[0] += s.v
	s.pad[1] += s.pad[0]
	s.pad[2] += s.pad[1]
	s.v += s.pad[2]
	s.pad[0] ^= s.v
	s.pad[1] ^= s.pad[0]
	s.pad[2] ^= s.pad[1]
	s.v ^= s.pad[2]
	s.pad[0] -= s.v
	s.pad[1] -= s.pad[0]
	s.pad[2] -= s.pad[1]
	s.v -= s.pad[2]
	s.pad[0] *= 3
	s.pad[1] *= 5
	s.pad[2] *= 7
	s.v *= 9
	return s
}

// DropSmall and DropSmallFat are the same call site twice. The only difference
// is which constructor they call, and therefore whether the compiler could see
// through it.
func DropSmall(v int) int    { return makeSmall(v).v }
func DropSmallFat(v int) int { return makeSmallFat(v).v }
