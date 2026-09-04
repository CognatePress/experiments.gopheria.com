// Package remote calls the constructor from the other side of a package
// boundary, which is the case the folklore says is different.
package remote

import "gopheria.lab/callsite"

// Drop is byte-for-byte the same call site as callsite.Drop, moved into another
// package. Export data carries New's body across the boundary, so whether this
// differs from the same-package version is a measurement rather than a
// prediction.
func Drop(v int) int {
	t := callsite.New(v)
	return t.V
}

// Keep is the escaping call site, from across the boundary.
func Keep(v int) int {
	t := callsite.New(v)
	callsite.Sink = t
	return t.V
}
