package callsite_test

import (
	"testing"

	"gopheria.lab/callsite"
	"gopheria.lab/callsite/remote"
)

var sink int

// BenchmarkWarmup runs first and is discarded — slot 02's finding, applied.
func BenchmarkWarmup(b *testing.B) {
	for b.Loop() {
		sink = callsite.Drop(1)
	}
}

// BenchmarkCallSite puts the same constructor's cost next to itself from five
// places. -m prints one verdict for New; this table prints five.
//
//	go test -run=^$ -bench='Warmup|CallSite' -benchmem -count=10 ./callsite |
//	  grep -v Warmup | benchstat -col /site -
func BenchmarkCallSite(b *testing.B) {
	for _, s := range []struct {
		name string
		fn   func(int) int
	}{
		{"site=same-package-drop", callsite.Drop},
		{"site=same-package-keep", callsite.Keep},
		{"site=same-package-sealed", callsite.Sealed},
		{"site=cross-package-drop", remote.Drop},
		{"site=cross-package-keep", remote.Keep},
	} {
		b.Run(s.name, func(b *testing.B) {
			b.ReportAllocs()
			n := 0
			for b.Loop() {
				n = s.fn(n + 1)
			}
			sink = n
		})
	}
}
