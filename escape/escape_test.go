package escape

import "testing"

var sinkInt int
var sinkByte byte
var sinkPtr *Point

func BenchmarkNewPoint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkPtr = NewPoint(i, i)
	}
}

func BenchmarkSum(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkInt = Sum(i, i)
	}
}

func BenchmarkFixedBuffer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkByte = FixedBuffer()
	}
}

func BenchmarkVarBufferSmall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkByte = VarBuffer(1024)
	}
}

func BenchmarkVarBufferLarge(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkByte = VarBuffer(100000)
	}
}

func BenchmarkSmallFixed(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkByte = SmallFixed()
	}
}

func BenchmarkLargeFixed(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkByte = LargeFixed()
	}
}

func BenchmarkSumAll(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkInt = SumAll(i, i, i)
	}
}
