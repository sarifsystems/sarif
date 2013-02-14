package stark

import (
	"testing"
)

func TestPath(t *testing.T) {
	str := "first/second/third"

	p := ParsePath(str)
	if p.Len() != 3 {
		t.Errorf("length: expected 3, got %d", p.Len())
	}

	if p.String() != str {
		t.Errorf("string: expected %v, got %v", str, p)
	}

	if p.First() != "first" {
		t.Errorf("first: expected %v, got %v", "first", p.First())
	}

	if p.Last() != "third" {
		t.Errorf("last: expected %v, got %v", "third", p.Last())
	}
}

func TestPathPop(t *testing.T) {
	str := "first/second/third"

	p := ParsePath(str)
	hop := p.Pop()
	if p.Len() != 2 {
		t.Errorf("length: expected 2, got %d", p.Len())
	}
	if hop != "first" {
		t.Errorf("hop: expected %v, got %v", "first", hop)
	}
	if p.String() != "second/third" {
		t.Errorf("string: expected %v, got %v", "second/third", p)
	}
	for i := 0; i < p.Len()+1; i++ {
		p.Pop()
	}
	if p.String() != "" {
		t.Errorf("string: expected %v, got %v", "", p)
	}
}

func TestPathPush(t *testing.T) {
	str := "first/second/third"

	p := ParsePath(str)
	p.Push("pushed")
	if p.Len() != 4 {
		t.Errorf("length: expected 4, got %d", p.Len())
	}
	if p.String() != "pushed/first/second/third" {
		t.Errorf("string: expected %v, got %v", "pushed/first/second/third", p)
	}
}

func TestCompleteRoute(t *testing.T) {
	r := ParseRoute("last/sender", "next/receiver")
	r.Forward("next")
	src, dest := r.Strings()
	if src != "next/last/sender" {
		t.Errorf("forward: expected next/last/sender, got %v", src)
	}
	if dest != "receiver" {
		t.Errorf("forward: expected receiver, got %v", dest)
	}
}

func TestIncompleteRoute(t *testing.T) {
	r := ParseRoute("last/sender", "receiver")
	r.Forward("next")
	src, dest := r.Strings()
	if src != "next/last/sender" {
		t.Errorf("forward: expected next/last/sender, got %v", src)
	}
	if dest != "receiver" {
		t.Errorf("forward: expected receiver, got %v", dest)
	}
}
