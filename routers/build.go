package routers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/labstack/echo/v4"
)

// braces renders :name parameters as {name}. net/http's pattern syntax and chi
// both use braces; gin and echo take the colon form unchanged.
func braces(pattern string) string {
	segs := strings.Split(pattern, "/")
	for i, s := range segs {
		if strings.HasPrefix(s, ":") {
			segs[i] = "{" + s[1:] + "}"
		}
	}
	return strings.Join(segs, "/")
}

// noop is the handler every route in every router carries. Dispatch is what is
// being measured, so the handler has to cost as close to nothing as a function
// call can.
func noop(http.ResponseWriter, *http.Request) {}

// NewStdlib builds net/http's ServeMux using the method-and-pattern syntax
// added in Go 1.22.
func NewStdlib(rs []Route) http.Handler {
	mux := http.NewServeMux()
	for _, r := range rs {
		mux.HandleFunc(r.Method+" "+braces(r.Pattern), noop)
	}
	return mux
}

// NewChi builds a chi router with no middleware stack. chi.NewRouter installs
// none by default; chi.NewMux with Use(...) is the shape most projects ship,
// and every entry in it would be measured as dispatch here.
func NewChi(rs []Route) http.Handler {
	mux := chi.NewRouter()
	for _, r := range rs {
		mux.Method(r.Method, braces(r.Pattern), http.HandlerFunc(noop))
	}
	return mux
}

// NewGin builds a gin engine with gin.New rather than gin.Default. Default
// installs the Logger and Recovery middleware, and a logger that formats a line
// per request would be the largest number in this table by a wide margin.
func NewGin(rs []Route) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	for _, r := range rs {
		e.Handle(r.Method, r.Pattern, func(*gin.Context) {})
	}
	return e
}

// NewEcho builds an echo instance, which ships with no middleware installed.
func NewEcho(rs []Route) http.Handler {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	for _, r := range rs {
		e.Add(r.Method, r.Pattern, func(echo.Context) error { return nil })
	}
	return e
}

// discard is a ResponseWriter that keeps nothing. httptest.NewRecorder
// allocates a body buffer and a header map per request, and that cost would
// show up in every row as the recorder's rather than the router's.
type discard struct{ h http.Header }

func newDiscard() *discard { return &discard{h: make(http.Header)} }

func (d *discard) Header() http.Header         { return d.h }
func (d *discard) Write(p []byte) (int, error) { return len(p), nil }
func (d *discard) WriteHeader(int)             {}
