package bce

import (
	"fmt"
	"testing"
)

var (
	data = func() []uint64 {
		xs := make([]uint64, 4096)
		for i := range xs {
			xs[i] = uint64(i) * 2654435761
		}
		return xs
	}()
	sink uint64
)

// BenchmarkWarmup runs first and is discarded — slot 02's finding, applied.
func BenchmarkWarmup(b *testing.B) {
	for b.Loop() {
		sink = Indexed(data[:64])
	}
}

// BenchmarkSum prices the five spellings at three lengths. The short length is
// where a per-iteration check would be hidden by loop setup; the long one is
// where it has room to show.
//
//	go test -run=^$ -bench='Warmup|Sum' -benchmem -count=10 ./bce |
//	  grep -v Warmup | benchstat -row /n -col /form -
func BenchmarkSum(b *testing.B) {
	forms := []struct {
		name string
		fn   func([]uint64) uint64
	}{
		{"form=indexed", Indexed},
		// A second copy of the baseline: same function, different row. Its true
		// delta against form=indexed is zero, so what benchstat reports for it
		// is this session's error bar — the control pair from slot 02.
		{"form=indexed-control", Indexed},
		{"form=ranged", Ranged},
		{"form=ranged-value", RangedValue},
		{"form=hinted", Hinted},
		{"form=bounded", func(s []uint64) uint64 { return Bounded(s, len(s)) }},
	}
	for _, f := range forms {
		b.Run(f.name, func(b *testing.B) {
			for _, n := range []int{64, 1024, 4096} {
				b.Run(fmt.Sprintf("n=%04d", n), func(b *testing.B) {
					s := data[:n]
					b.ReportAllocs()
					for b.Loop() {
						sink = f.fn(s)
					}
				})
			}
		})
	}
}
