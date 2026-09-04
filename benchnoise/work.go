// Package benchnoise carries two workloads whose relative cost is known by
// construction, so that a benchmark's reported delta can be compared against a
// ground truth instead of against another benchmark.
package benchnoise

const (
	sumN    = 1000
	sumN5   = 1050 // 5% more additions than sumN
	allocN  = 200
	allocN5 = 210 // 5% more allocations than allocN
)

var data = func() []uint64 {
	xs := make([]uint64, sumN5)
	for i := range xs {
		xs[i] = uint64(i) * 2654435761
	}
	return xs
}()

var (
	sinkU64 uint64
	sinkB   []byte
	sinkInt int
)

// Sum is compute-bound and allocates nothing. Cost is linear in len(xs), which
// is what makes a 5% longer slice a 5% larger benchmark.
//
//go:noinline
func Sum(xs []uint64) uint64 {
	var s uint64
	for _, v := range xs {
		s += v
	}
	return s
}

// Alloc allocates n 64-byte slices on the heap. The assignment to sinkB is what
// puts them there: without it escape analysis keeps every one on the stack and
// the benchmark measures nothing.
//
//go:noinline
func Alloc(n int) int {
	total := 0
	for i := 0; i < n; i++ {
		b := make([]byte, 64)
		b[0] = byte(i)
		sinkB = b
		total += len(b)
	}
	return total
}
