package inlinebudget

import "testing"

var out int

// BenchmarkWarmup runs first and is discarded — slot 02's finding, applied.
func BenchmarkWarmup(b *testing.B) {
	for b.Loop() {
		out = Grow01(out)
	}
}

// BenchmarkBoundary prices the inlining budget's edge. Grow26 costs exactly 80
// and is accepted; Grow27 costs 83 and is refused. They differ by one `n += k`
// statement, so everything except the inlining decision is held constant.
//
// grow26-control is a second row for Grow26, so the harness's own error bar is
// printed next to a difference that is expected to be small.
//
//	go test -run=^$ -bench='Warmup|Boundary' -benchmem -count=10 ./inlinebudget |
//	  grep -v Warmup | benchstat -col /fn -
func BenchmarkBoundary(b *testing.B) {
	for _, f := range []struct {
		name string
		fn   func(int) int
	}{
		{"fn=grow26-inlined", Grow26},
		{"fn=grow26-control", Grow26},
		{"fn=grow27-not-inlined", Grow27},
	} {
		b.Run(f.name, func(b *testing.B) {
			b.ReportAllocs()
			n := 0
			for b.Loop() {
				n = f.fn(n)
			}
			out = n
		})
	}
}

// BenchmarkDirect calls the same two functions by name rather than through a
// func value. The indirect call above cannot be inlined whatever the budget
// says, so it measures the two bodies with the decision removed; this one
// measures the decision.
func BenchmarkDirect(b *testing.B) {
	b.Run("fn=grow26-inlined", func(b *testing.B) {
		n := 0
		for b.Loop() {
			n = Grow26(n)
		}
		out = n
	})
	b.Run("fn=grow26-control", func(b *testing.B) {
		n := 0
		for b.Loop() {
			n = Grow26(n)
		}
		out = n
	})
	b.Run("fn=grow27-not-inlined", func(b *testing.B) {
		n := 0
		for b.Loop() {
			n = Grow27(n)
		}
		out = n
	})
}

// BenchmarkEscape is the cost of the budget decision, as opposed to the cost of
// the call. Both rows construct a value and drop it; only one of the two
// constructors fits in the budget, and the other therefore cannot be seen
// through — so the pointer it returns has to be assumed to escape.
//
//	go test -run=^$ -bench='Warmup|Escape' -benchmem -count=10 ./inlinebudget |
//	  grep -v Warmup | benchstat -col /ctor -
func BenchmarkEscape(b *testing.B) {
	for _, c := range []struct {
		name string
		fn   func(int) int
	}{
		{"ctor=inlinable", DropSmall},
		{"ctor=inlinable-control", DropSmall},
		{"ctor=over-budget", DropSmallFat},
	} {
		b.Run(c.name, func(b *testing.B) {
			b.ReportAllocs()
			n := 0
			for b.Loop() {
				n += c.fn(n)
			}
			out = n
		})
	}
}
