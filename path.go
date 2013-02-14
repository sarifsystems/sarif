package stark

import (
	"strings"
)

// The separator between hops in a path.
const PathSep string = "/"

// Path provides convenience functions for manipulating message paths.
type Path struct {
	Parts []string
}

// ParsePath parses a path string into a Path struct.
// A path, as defined in the spec, is a string of zero or more named hops
// (services) separated by the path separator "/", e.g. "next/second/last".
func ParsePath(path string) *Path {
	return &Path{
		strings.Split(path, PathSep),
	}
}

// String returns a spec-conform string representation of the path.
func (p *Path) String() string {
	return strings.Join(p.Parts, PathSep)
}

// First returns the next hop in the path or an empty string if the path has no hops.
func (p *Path) First() string {
	if len(p.Parts) == 0{
		return ""
	}
	return p.Parts[0]
}

// Len returns the number of hops in this path.
func (p *Path) Len() int {
	return len(p.Parts)
}

// Last returns the last hop in the path or an empty string if the path has no hops.
func (p *Path) Last() string {
	l := len(p.Parts)
	if l == 0 {
		return ""
	}
	return p.Parts[l-1]
}

// Push pushes a new hop at the beginning of the path so that it becomes the first one.
func (p *Path) Push(hop string) {
	p.Parts = append([]string{hop}, p.Parts...)
}

// Pop removes the first hop from the path and returns it.
func (p *Path) Pop() string {
	if len(p.Parts) == 0 {
		return ""
	}
	next := p.Parts[0]
	p.Parts = p.Parts[1:]
	return next
}

// Route provides a wrapper around both source and destination paths of a message.
type Route struct {
	Source *Path
	Dest *Path
}

// ParseRoute parses source and destination paths into a new route.
func ParseRoute(src, dest string) *Route {
	return &Route{
		ParsePath(src),
		ParsePath(dest),
	}
}

// String returns a human-readable form of the route.
func (r *Route) String() string {
	src, dest := r.Strings()
	return src + " -> " + dest
}

// Strings returns spec-conform source and destination paths.
func (r *Route) Strings() (src, dest string) {
	return r.Source.String(), r.Dest.String()
}

// Forward moves the route one hop along the path, ideally popping the next hop
// from the destination and pushing it on the source. The passed hop specifies the
// current hop being moved. If it differs from the next hop, it is inserted instead.
//
// For example, when calling Forward("next"):
// "last -> next/dest" and "last -> dest" both become "next/last -> "dest"
func (r *Route) Forward(hop string) string {
	if hop == r.Dest.First() {
		r.Dest.Pop()
	}
	r.Source.Push(hop)
	return r.Dest.First()
}
