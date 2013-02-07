package stark

import (
	"strings"
)

const PATH_SEPARATOR string = "/"

type Path struct {
	source []string
	destination []string
}

func NewPath(src, dest string) *Path {
	return &Path{
		strings.Split(src, PATH_SEPARATOR),
		strings.Split(dest, PATH_SEPARATOR),
	}
}

func GetPath(m *Message) *Path {
	return NewPath(m.Source, m.Destination)
}

func (p *Path) Next() string {
	return p.destination[0]
}

func (p *Path) Previous() string {
	return p.source[0]
}

func (p *Path) Source() string {
	return strings.Join(p.source, PATH_SEPARATOR)
}

func (p *Path) Destination() string {
	return strings.Join(p.destination, PATH_SEPARATOR)
}

func (p *Path) Sender() string {
	return p.source[len(p.source)-1]
}

func (p *Path) Receiver() string {
	return p.destination[len(p.destination)-1]
}

func (p *Path) Apply(m *Message) {
	m.Source = p.Source()
	m.Destination = p.Destination()
}
