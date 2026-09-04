// Package callsite carries one constructor and several places that call it.
//
// L2's FAQ left a claim standing: -gcflags=-m reports a verdict per function,
// while allocation happens per call site. The same New below allocates in some
// of the callers here and not in others, and -m prints the same line for all of
// them.
package callsite

// T is small enough that the compiler will consider keeping it in a frame, and
// large enough that its allocation is visible in -benchmem.
type T struct {
	V   int
	buf [3]int
}

// New is the constructor under test. Its body is well inside the inlining
// budget, which is the point — the escape verdict printed for New is fixed,
// and what varies is whether the caller inlined it.
func New(v int) *T { return &T{V: v} }

// NewSealed is the same constructor with inlining disabled. A caller cannot see
// through it, so the returned pointer escapes by definition: the compiler has
// to assume the caller keeps it.
//
//go:noinline
func NewSealed(v int) *T { return &T{V: v} }

// Sink is where the escaping call site puts its result.
var Sink *T

// Drop calls New and lets the value die in this frame.
func Drop(v int) int {
	t := New(v)
	return t.V
}

// Keep calls New and stores the pointer beyond the frame.
func Keep(v int) int {
	t := New(v)
	Sink = t
	return t.V
}

// Sealed calls the non-inlinable copy and lets the value die — which it cannot,
// because the compiler compiled New's body without knowing that.
func Sealed(v int) int {
	t := NewSealed(v)
	return t.V
}
