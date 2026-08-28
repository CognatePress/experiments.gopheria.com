package escape

import "fmt"

type Point struct{ X, Y int }

// 1. The textbook case: the pointer outlives the frame.
func NewPoint(x, y int) *Point {
	p := Point{x, y}
	return &p
}

// 2. The address is taken, but never leaves. Intuition says heap; it is not.
func Sum(x, y int) int {
	p := Point{x, y}
	q := &p
	return q.X + q.Y
}

// 3. Nothing here outlives the call, yet the argument escapes anyway.
func Log(x int) {
	fmt.Fprint(discard{}, x)
}

type discard struct{}

func (discard) Write(b []byte) (int, error) { return len(b), nil }

// 4. Same allocation, one known at compile time and one not.
func FixedBuffer() byte {
	buf := make([]byte, 1024)
	return buf[0]
}

func VarBuffer(n int) byte {
	buf := make([]byte, n)
	return buf[0]
}

// 5. Both sizes are constant. Only one of them fits.
func SmallFixed() byte {
	buf := make([]byte, 65536)
	return buf[0]
}

func LargeFixed() byte {
	buf := make([]byte, 65537)
	return buf[0]
}

// 6. A local slice of pointers to locals — no escape, despite the indirection.
func SumAll(a, b, c int) int {
	x, y, z := a, b, c
	ps := []*int{&x, &y, &z}
	total := 0
	for _, p := range ps {
		total += *p
	}
	return total
}

// 7. The closure does not escape; the goroutine's copy of it does.
func Local(x int) int {
	f := func() int { return x * 2 }
	return f()
}

func Spawned(x int, done chan int) {
	go func() { done <- x * 2 }()
}
