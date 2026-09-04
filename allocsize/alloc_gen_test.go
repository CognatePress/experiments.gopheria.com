package allocsize

import (
	"fmt"
	"testing"
	"unsafe"
)

// The types below are pointer-containing, so they take the runtime's scanning
// allocation path. The []byte benchmarks alongside them take the pointer-free
// one. Both are here because the two paths are specialised separately, and a
// table with only one of them would generalise from half the change.

type p8 struct {
	p *byte
}
type p16 struct {
	p *byte
	_ [8]byte
}
type p24 struct {
	p *byte
	_ [16]byte
}
type p32 struct {
	p *byte
	_ [24]byte
}
type p48 struct {
	p *byte
	_ [40]byte
}
type p64 struct {
	p *byte
	_ [56]byte
}
type p80 struct {
	p *byte
	_ [72]byte
}
type p96 struct {
	p *byte
	_ [88]byte
}
type p112 struct {
	p *byte
	_ [104]byte
}
type p128 struct {
	p *byte
	_ [120]byte
}
type p160 struct {
	p *byte
	_ [152]byte
}
type p256 struct {
	p *byte
	_ [248]byte
}
type p512 struct {
	p *byte
	_ [504]byte
}
type p1024 struct {
	p *byte
	_ [1016]byte
}

// BenchmarkWarmup runs first and is discarded — slot 02's finding, applied.
func BenchmarkWarmup(b *testing.B) {
	for b.Loop() {
		sinkB = make([]byte, 64)
	}
}

// BenchmarkAlloc is the sweep. Each sub-benchmark allocates one object of one
// constant size, on one of the two allocation paths.
//
//	go test -run=^$ -bench='Warmup|Alloc' -benchmem -count=10 ./allocsize |
//	  grep -v Warmup | benchstat -row /size -col /shape -
func BenchmarkAlloc(b *testing.B) {
	b.Run("shape=noscan", func(b *testing.B) {
		b.Run(name(8), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkB = make([]byte, 8)
			}
		})
		b.Run(name(16), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkB = make([]byte, 16)
			}
		})
		b.Run(name(24), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkB = make([]byte, 24)
			}
		})
		b.Run(name(32), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkB = make([]byte, 32)
			}
		})
		b.Run(name(48), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkB = make([]byte, 48)
			}
		})
		b.Run(name(64), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkB = make([]byte, 64)
			}
		})
		b.Run(name(80), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkB = make([]byte, 80)
			}
		})
		b.Run(name(96), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkB = make([]byte, 96)
			}
		})
		b.Run(name(112), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkB = make([]byte, 112)
			}
		})
		b.Run(name(128), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkB = make([]byte, 128)
			}
		})
		b.Run(name(160), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkB = make([]byte, 160)
			}
		})
		b.Run(name(256), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkB = make([]byte, 256)
			}
		})
		b.Run(name(512), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkB = make([]byte, 512)
			}
		})
		b.Run(name(1024), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				sinkB = make([]byte, 1024)
			}
		})
	})

	b.Run("shape=scan", func(b *testing.B) {
		b.Run(name(8), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				keep(unsafe.Pointer(new(p8)))
			}
		})
		b.Run(name(16), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				keep(unsafe.Pointer(new(p16)))
			}
		})
		b.Run(name(24), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				keep(unsafe.Pointer(new(p24)))
			}
		})
		b.Run(name(32), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				keep(unsafe.Pointer(new(p32)))
			}
		})
		b.Run(name(48), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				keep(unsafe.Pointer(new(p48)))
			}
		})
		b.Run(name(64), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				keep(unsafe.Pointer(new(p64)))
			}
		})
		b.Run(name(80), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				keep(unsafe.Pointer(new(p80)))
			}
		})
		b.Run(name(96), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				keep(unsafe.Pointer(new(p96)))
			}
		})
		b.Run(name(112), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				keep(unsafe.Pointer(new(p112)))
			}
		})
		b.Run(name(128), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				keep(unsafe.Pointer(new(p128)))
			}
		})
		b.Run(name(160), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				keep(unsafe.Pointer(new(p160)))
			}
		})
		b.Run(name(256), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				keep(unsafe.Pointer(new(p256)))
			}
		})
		b.Run(name(512), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				keep(unsafe.Pointer(new(p512)))
			}
		})
		b.Run(name(1024), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				keep(unsafe.Pointer(new(p1024)))
			}
		})
	})
}

// name formats the sub-benchmark label so benchstat can pivot on it. The
// zero-padding keeps the rows in numeric order in the output table.
func name(size int) string { return fmt.Sprintf("size=%04d", size) }
