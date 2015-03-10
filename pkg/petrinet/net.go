// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package petrinet

import (
	"fmt"
	"strings"
)

type Node struct {
	Name string
	In   []*Transition
	Out  []*Transition

	Tokens   int
	OnChange func(name string, prev, curr int)
}

func (n Node) String() string {
	return n.Name
}

func (n *Node) AddToken(i int) {
	n.Tokens += i
	if i > 0 {
		for _, t := range n.Out {
			if !t.CouldFire && t.CanFire() {
				t.CouldFire = true
			}
		}
	} else {
		for _, t := range n.Out {
			if t.CouldFire && !t.CanFire() {
				t.CouldFire = false
			}
		}
	}
	if n.OnChange != nil {
		n.OnChange(n.Name, n.Tokens-i, n.Tokens)
	}
}

type Transition struct {
	Name      string
	CouldFire bool
	In        []*Node
	Out       []*Node
}

func (t Transition) String() string {
	var in, out []string
	for _, node := range t.In {
		in = append(in, node.Name)
	}
	for _, node := range t.Out {
		out = append(out, node.Name)
	}

	return fmt.Sprintf("[%s] => [%s]", strings.Join(in, ", "), strings.Join(out, ", "))
}

func (t *Transition) CanFire() bool {
	for _, n := range t.In {
		if n.Tokens < 1 {
			return false
		}
	}
	return true
}

func (t *Transition) Fire() {
	if t.CouldFire == false {
		return
	}
	for _, n := range t.In {
		n.AddToken(-1)
	}
	for _, n := range t.Out {
		n.AddToken(1)
	}
}

type Net struct {
	Nodes       map[string]*Node
	Transitions []*Transition
}

func New() *Net {
	return &Net{
		make(map[string]*Node),
		make([]*Transition, 0),
	}
}

func (n *Net) AddTransition(in, out []string) *Transition {
	t := &Transition{}
	n.Transitions = append(n.Transitions, t)
	for _, name := range in {
		node := n.GetNode(name)
		t.In = append(t.In, node)
		node.Out = append(node.Out, t)
	}
	for _, name := range out {
		node := n.GetNode(name)
		t.Out = append(t.Out, node)
		node.In = append(node.In, t)
	}
	t.CouldFire = t.CanFire()
	return t
}

func (n *Net) GetNode(name string) *Node {
	if node, ok := n.Nodes[name]; ok {
		return node
	}
	node := &Node{
		Name: name,
	}
	n.Nodes[name] = node
	return node
}

func (n *Net) Spawn(name string, i int) {
	node, ok := n.Nodes[name]
	if !ok {
		return
	}
	node.AddToken(i)
}

func (n *Net) Run(iterations int) bool {
STEP:
	for {
		if iterations == 0 {
			return false
		}
		iterations--
		for _, t := range n.Transitions {
			if t.CouldFire {
				t.Fire()
				continue STEP
			}
		}
		return true
	}
}
