// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package petrinet

import "testing"

func TestSimpleFiring(t *testing.T) {
	n := New()
	n.AddTransition([]string{"input/node"}, []string{"output/node"})

	fired := 0
	n.GetNode("output/node").OnChange = func(name string, prev, curr int) {
		if name != "output/node" {
			t.Fatal("wrong name:", name)
		}
		if prev != fired {
			t.Fatalf("expected %d prev tokens, got %d", fired, prev)
		}
		fired++
		if curr != fired {
			t.Fatalf("expected %d curr tokens, got %d", fired, curr)
		}
	}

	n.Spawn("input/node", 1)
	n.Run(-1)
	if fired != 1 {
		t.Fatalf("expected output/node to fire %d times, but got %d", 1, fired)
	}
}

func TestComplexFiring(t *testing.T) {
	n := New()
	n.AddTransition([]string{"input/single"}, []string{"output/multi1", "output/multi2"})
	n.AddTransition([]string{"input/multi1", "input/multi2"}, []string{"output/single"})

	total := 0
	fired := make(map[string]int)
	for _, node := range n.Nodes {
		node.OnChange = func(name string, prev, curr int) {
			if curr > prev {
				total++
				fired[name] += (curr - prev)
			}
		}
	}

	n.Run(-1)
	if exp := 0; total != exp {
		t.Log(fired)
		t.Fatalf("expected %d firings, got %d", exp, total)
	}

	// test multi output
	n.Spawn("input/single", 1)
	n.Run(-1)
	if exp := 3; total != exp {
		t.Log(fired)
		t.Fatalf("expected %d firings, got %d", exp, total)
	}
	if v := fired["output/multi1"]; v != 1 {
		t.Fatalf("output/multi1 should fire")
	}
	if v := fired["output/multi2"]; v != 1 {
		t.Fatalf("output/multi2 should fire")
	}

	// test multi input
	total = 0
	fired = make(map[string]int)
	n.Spawn("input/multi1", 1)
	n.Run(-1)
	if exp := 1; total != exp {
		t.Log(fired)
		t.Fatalf("expected %d firings, got %d", exp, total)
	}

	n.Spawn("input/multi2", 1)
	n.Run(-1)
	if exp := 3; total != exp {
		t.Log(fired)
		t.Fatalf("expected %d firings, got %d", exp, total)
	}
	if v := fired["output/single"]; v != 1 {
		t.Fatalf("output/single should fire")
	}

	// test cascading
	n.AddTransition([]string{"output/multi1"}, []string{"input/multi1"})
	n.GetNode("output/multi1").Tokens = 0
	total = 0
	fired = make(map[string]int)
	n.Spawn("input/single", 1)
	n.Spawn("input/multi2", 1)
	n.Run(-1)
	if exp := 6; total != exp {
		t.Log(fired)
		t.Fatalf("expected %d firings, got %d", exp, total)
	}
}

func TestInfinite(t *testing.T) {
	n := New()
	n.AddTransition([]string{"one"}, []string{"two"})
	n.AddTransition([]string{"two"}, []string{"one"})

	fired := 0
	n.GetNode("one").OnChange = func(name string, prev, curr int) {
		if curr > prev {
			fired++
		}
	}

	n.Run(-1)

	n.Spawn("one", 1)
	n.Run(1000)
	if exp := 501; fired != exp {
		t.Log(fired)
		t.Fatalf("expected %d firings, got %d", exp, fired)
	}
}
