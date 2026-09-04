package benchnoise

import (
	"flag"
	"testing"
)

// order controls the sequence the three implementations run in. Everything else
// about the run is identical, so any delta that changes sign when this flag
// changes is a property of the harness, not of the code being measured.
var order = flag.String("order", "forward", "implementation order: forward or reverse")

// BenchmarkWarmup runs before BenchmarkAB and its result is discarded. It is
// here to test whether the penalty the first benchmark in a run pays can be
// spent on a benchmark nobody reads.
func BenchmarkWarmup(b *testing.B) {
	xs := data[:sumN]
	b.ReportAllocs()
	for b.Loop() {
		sinkU64 = Sum(xs)
	}
}

type impl struct {
	name string
	run  func(*testing.B)
}

// BenchmarkAB runs three implementations of two workloads. alpha and beta call
// the same function with the same argument, so their true difference is zero.
// plus5 does 5% more work than alpha, so its true difference is +5%.
func BenchmarkAB(b *testing.B) {
	for _, w := range []struct {
		name  string
		impls []impl
	}{
		{"work=sum", []impl{
			{"impl=alpha", func(b *testing.B) { benchSum(b, sumN) }},
			{"impl=beta", func(b *testing.B) { benchSum(b, sumN) }},
			{"impl=plus5", func(b *testing.B) { benchSum(b, sumN5) }},
		}},
		{"work=alloc", []impl{
			{"impl=alpha", func(b *testing.B) { benchAlloc(b, allocN) }},
			{"impl=beta", func(b *testing.B) { benchAlloc(b, allocN) }},
			{"impl=plus5", func(b *testing.B) { benchAlloc(b, allocN5) }},
		}},
	} {
		impls := w.impls
		if *order == "reverse" {
			impls = reversed(impls)
		}
		b.Run(w.name, func(b *testing.B) {
			for _, im := range impls {
				b.Run(im.name, im.run)
			}
		})
	}
}

func reversed(in []impl) []impl {
	out := make([]impl, len(in))
	for i, im := range in {
		out[len(in)-1-i] = im
	}
	return out
}

func benchSum(b *testing.B, n int) {
	xs := data[:n]
	b.ReportAllocs()
	for b.Loop() {
		sinkU64 = Sum(xs)
	}
}

func benchAlloc(b *testing.B, n int) {
	b.ReportAllocs()
	for b.Loop() {
		sinkInt = Alloc(n)
	}
}
