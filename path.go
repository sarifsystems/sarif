package stark

import (
	"strings"
)

const PathSep string = "/"

type Path struct {
	Parts []string
}

func ParsePath(path string) *Path {
	return &Path{
		strings.Split(path, PathSep),
	}
}

func (p *Path) String() string {
	return strings.Join(p.Parts, PathSep)
}

func (p *Path) First() string {
	if len(p.Parts) == 0{
		return ""
	}
	return p.Parts[0]
}

func (p *Path) Len() int {
	return len(p.Parts)
}

func (p *Path) Last() string {
	l := len(p.Parts)
	if l == 0 {
		return ""
	}
	return p.Parts[l-1]
}

func (p *Path) Push(hop string) {
	p.Parts = append([]string{hop}, p.Parts...)
}

func (p *Path) Pop() string {
	if len(p.Parts) == 0 {
		return ""
	}
	next := p.Parts[0]
	p.Parts = p.Parts[1:]
	return next
}

type Route struct {
	Source *Path
	Dest *Path
}

func ParseRoute(src, dest string) *Route {
	return &Route{
		ParsePath(src),
		ParsePath(dest),
	}
}

func (r *Route) String() string {
	src, dest := r.Strings()
	return src + " -> " + dest
}

func (r *Route) Strings() (string, string) {
	return r.Source.String(), r.Dest.String()
}

func (r *Route) Forward(hop string) string {
	if hop == r.Dest.First() {
		r.Dest.Pop()
	}
	r.Source.Push(hop)
	return r.Dest.First()
}
