package routers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// BenchmarkWarmup runs before the others and its result is discarded — the
// first benchmark in a process pays a penalty the rest do not, and slot 02's
// spread runs showed it is large enough to change a verdict.
func BenchmarkWarmup(b *testing.B) {
	h := NewStdlib(Table())
	req := httptest.NewRequest(http.MethodGet, StaticPath, nil)
	w := newDiscard()
	for b.Loop() {
		h.ServeHTTP(w, req)
	}
}

// BenchmarkDispatch is the table: four routers, one 200-route table, three
// requests each.
//
// The request is built once and reused. Every router here either leaves it
// untouched or replaces the fields it needs on each call, so reuse costs
// nothing that a fresh request would not — and building one inside the loop
// would put httptest.NewRequest's allocations in every row.
//
//	go test -run=^$ -bench='Warmup|Dispatch' -benchmem -count=10 ./routers |
//	  grep -v Warmup | benchstat -row /path -col /router -
func BenchmarkDispatch(b *testing.B) {
	rs := Table()
	routers := []struct {
		name string
		h    http.Handler
	}{
		{"router=stdlib", NewStdlib(rs)},
		// A second copy of the baseline. Its true delta against router=stdlib
		// is zero, so whatever benchstat reports for it is this session's error
		// bar — the control pair from slot 02.
		{"router=stdlib-control", NewStdlib(rs)},
		{"router=chi", NewChi(rs)},
		{"router=gin", NewGin(rs)},
		{"router=echo", NewEcho(rs)},
	}
	paths := []struct{ name, method, path string }{
		{"path=static", http.MethodGet, StaticPath},
		{"path=param", http.MethodGet, ParamPath},
		{"path=miss", http.MethodGet, MissPath},
	}

	for _, rt := range routers {
		b.Run(rt.name, func(b *testing.B) {
			for _, p := range paths {
				b.Run(p.name, func(b *testing.B) {
					req := httptest.NewRequest(p.method, p.path, nil)
					w := newDiscard()
					b.ReportAllocs()
					for b.Loop() {
						rt.h.ServeHTTP(w, req)
					}
				})
			}
		})
	}
}

// session is a small response body — five fields, one of them a time — of the
// kind a handler behind any of these routes would return.
type session struct {
	ID        string    `json:"id"`
	User      string    `json:"user"`
	Org       string    `json:"org"`
	CreatedAt time.Time `json:"created_at"`
	Active    bool      `json:"active"`
}

var sink []byte

// BenchmarkEncode is the ruler the dispatch numbers are read against: one
// json.Marshal of that body, which is the cheapest useful thing a handler can
// do after the router has finished with the request.
func BenchmarkEncode(b *testing.B) {
	s := session{
		ID:        "9f3c2a",
		User:      "nolan",
		Org:       "acme",
		CreatedAt: time.Date(2026, 9, 20, 9, 0, 0, 0, time.UTC),
		Active:    true,
	}
	b.ReportAllocs()
	for b.Loop() {
		out, err := json.Marshal(s)
		if err != nil {
			b.Fatal(err)
		}
		sink = out
	}
}

// BenchmarkMiss reads the cost of a request that matches nothing against the
// size of the table it failed to match in. A hit is answered by a tree walk
// whose cost is bounded by the path; a miss, in net/http's mux, is not.
//
//	go test -run=^$ -bench='Warmup|Miss' -benchmem -count=10 ./routers |
//	  grep -v Warmup | benchstat -row /routes -col /router -
func BenchmarkMiss(b *testing.B) {
	for _, n := range []int{2, 5, 10, 20} {
		rs := TableN(n)
		b.Run(fmt.Sprintf("routes=%03d", len(rs)), func(b *testing.B) {
			for _, rt := range []struct {
				name string
				h    http.Handler
			}{
				{"router=stdlib", NewStdlib(rs)},
				{"router=chi", NewChi(rs)},
				{"router=gin", NewGin(rs)},
				{"router=echo", NewEcho(rs)},
			} {
				b.Run(rt.name, func(b *testing.B) {
					req := httptest.NewRequest(http.MethodGet, MissPath, nil)
					w := newDiscard()
					b.ReportAllocs()
					for b.Loop() {
						rt.h.ServeHTTP(w, req)
					}
				})
			}
		})
	}
}

// BenchmarkMissMethods holds the paths fixed and varies how many distinct
// methods the table registers. If the miss cost is the per-method tree walk in
// ServeMux.matchingMethods, this is the axis it moves along — and the
// route-count sweep above is the axis it does not.
//
//	go test -run=^$ -bench='Warmup|MissMethods' -benchmem -count=10 ./routers |
//	  grep -v Warmup | benchstat -row /methods -col /router -
func BenchmarkMissMethods(b *testing.B) {
	full := Table()
	for k := 1; k <= len(Methods(full)); k++ {
		rs := KeepMethods(full, k)
		b.Run(fmt.Sprintf("methods=%d", k), func(b *testing.B) {
			b.Logf("%d routes across %d methods", len(rs), k)
			for _, rt := range []struct {
				name string
				h    http.Handler
			}{
				{"router=stdlib", NewStdlib(rs)},
				{"router=gin", NewGin(rs)},
			} {
				b.Run(rt.name, func(b *testing.B) {
					req := httptest.NewRequest(http.MethodGet, MissPath, nil)
					w := newDiscard()
					b.ReportAllocs()
					for b.Loop() {
						rt.h.ServeHTTP(w, req)
					}
				})
			}
		})
	}
}
