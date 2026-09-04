// Package bce holds four spellings of the same sum, so the compiler's own
// bounds-check accounting can be read against what each spelling costs.
//
// Every function returns the same value for the same input. The only thing that
// differs is how the loop is written, which is exactly what the folklore is
// about.
//
//	go build -gcflags='-d=ssa/check_bce/debug=1' ./bce
package bce

// Indexed is the plain form: the loop bound is len(s), re-read by the compiler
// at each comparison.
func Indexed(s []uint64) uint64 {
	var sum uint64
	for i := 0; i < len(s); i++ {
		sum += s[i]
	}
	return sum
}

// Ranged indexes inside a range loop. The index is provably in range because
// range produced it.
func Ranged(s []uint64) uint64 {
	var sum uint64
	for i := range s {
		sum += s[i]
	}
	return sum
}

// RangedValue never indexes at all.
func RangedValue(s []uint64) uint64 {
	var sum uint64
	for _, v := range s {
		sum += v
	}
	return sum
}

// Hinted is the folklore version. The discarded index at the top is the "BCE
// hint": it asserts the slice is at least len(s) long, which the compiler
// already knew.
func Hinted(s []uint64) uint64 {
	if len(s) == 0 {
		return 0
	}
	_ = s[len(s)-1]
	var sum uint64
	for i := 0; i < len(s); i++ {
		sum += s[i]
	}
	return sum
}

// Bounded takes the length as a separate argument, so the compiler cannot
// relate it to the slice. This is the case where a check genuinely survives,
// and it is here to price one.
func Bounded(s []uint64, n int) uint64 {
	var sum uint64
	for i := 0; i < n; i++ {
		sum += s[i]
	}
	return sum
}
