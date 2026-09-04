// Package routers measures what four Go routers charge to pick a handler, with
// the handler removed.
//
// Router benchmarks usually leave real work in the handler, which drowns the
// thing being measured. Here every handler is empty, the ResponseWriter writes
// nowhere, and the request is built once outside the timed region — so what is
// left is dispatch: matching the path, extracting the parameters, and whatever
// per-request bookkeeping the router does on the way.
package routers

import "fmt"

// Route is a router-agnostic pattern. A segment written :name is a parameter;
// each router's builder renders it in that router's own syntax.
type Route struct {
	Method  string
	Pattern string
}

// resources are the nouns the generated table is built from. Real route tables
// are shaped like this — a handful of collections, each with a similar set of
// verbs — rather than being 200 unrelated strings.
var resources = []string{
	"users", "accounts", "projects", "repos", "issues",
	"comments", "labels", "milestones", "releases", "tags",
	"builds", "artifacts", "runners", "secrets", "webhooks",
	"teams", "invitations", "policies", "audits", "sessions",
}

// Table returns 200 routes: ten per resource, mixing collection, item, nested
// and action shapes.
func Table() []Route { return TableN(len(resources)) }

// TableN returns the same shapes built from the first n resources, so the cost
// of a lookup can be read against the size of the table it searched. Ten routes
// per resource, so n=20 is the 200-route table above.
func TableN(n int) []Route {
	if n > len(resources) {
		n = len(resources)
	}
	rs := make([]Route, 0, n*10)
	for _, r := range resources[:n] {
		rs = append(rs,
			Route{"GET", "/api/v1/" + r},
			Route{"POST", "/api/v1/" + r},
			Route{"GET", "/api/v1/" + r + "/:id"},
			Route{"PUT", "/api/v1/" + r + "/:id"},
			Route{"DELETE", "/api/v1/" + r + "/:id"},
			Route{"GET", "/api/v1/" + r + "/:id/history"},
			Route{"GET", "/api/v1/orgs/:org/" + r},
			Route{"GET", "/api/v1/orgs/:org/" + r + "/:id"},
			Route{"POST", "/api/v1/orgs/:org/" + r + "/:id/archive"},
			// The nested prefix is "groups" rather than one of the resources
			// above. Gin's tree, inherited from httprouter, requires the same
			// wildcard name at a given position in a subtree, so reusing
			// "teams" here — a resource that already generates
			// /api/v1/orgs/:org/teams/:id — makes it panic at registration.
			Route{"GET", "/api/v1/orgs/:org/groups/:group/" + r + "/:id/status"},
		)
	}
	if len(rs) != n*10 {
		panic(fmt.Sprintf("route table is %d routes, want %d", len(rs), n*10))
	}
	return rs
}

// The three requests every router is measured against. They are picked to
// separate the costs a single average would hide: a static path that matches on
// literal segments alone, a deep path carrying three parameters, and a path
// that matches nothing.
const (
	StaticPath = "/api/v1/sessions"
	ParamPath  = "/api/v1/orgs/acme/groups/platform/sessions/9f3c2a/status"
	MissPath   = "/api/v1/orgs/acme/groups/platform/sessions/9f3c2a/nonesuch"
)

// Methods returns the distinct methods the table registers, in the order they
// first appear.
func Methods(rs []Route) []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range rs {
		if !seen[r.Method] {
			seen[r.Method] = true
			out = append(out, r.Method)
		}
	}
	return out
}

// KeepMethods returns the routes whose method is one of the first k the table
// registers. It exists to separate two things a plain route-count sweep cannot:
// net/http walks the routing tree once per registered method when a request
// matches nothing, so the miss cost should track the number of methods rather
// than the number of routes.
func KeepMethods(rs []Route, k int) []Route {
	keep := map[string]bool{}
	for _, m := range Methods(rs)[:k] {
		keep[m] = true
	}
	out := make([]Route, 0, len(rs))
	for _, r := range rs {
		if keep[r.Method] {
			out = append(out, r)
		}
	}
	return out
}
