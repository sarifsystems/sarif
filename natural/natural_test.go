package natural

import (
	"fmt"
	"testing"
)

func TestSplitActionWildcard(t *testing.T) {
	a, p, n := SplitActionWildcard("simple")
	if a != "simple" || p != 0 || n != 0 {
		t.Errorf("Expected %v,%v,%v, got %v,%v,%v", "simple", 0, 0, a, p, n)
	}

	a, p, n = SplitActionWildcard("***.testing.**")
	if a != "testing" || p != 3 || n != 2 {
		t.Errorf("Expected %v,%v,%v, got %v,%v,%v", "testing", 3, 2, a, p, n)
	}
}

func TestCombineAction(t *testing.T) {
	// Simple
	expect, got := "hello", CombineActions([]string{
		"hello",
		"music.*",
		"*.play",
		"*.pause",
	})
	if got != expect {
		t.Errorf("Expected %v, got %v", expect, got)
	}

	// Two-part
	expect, got = "music.play", CombineActions([]string{
		"*.play",
		"music.*",
		"*.pause",
	})
	if got != expect {
		t.Errorf("Expected %v, got %v", expect, got)
	}

	// Three-part
	expect, got = "first.second.third", CombineActions([]string{
		"first.**",
		"hello",
		"*.second.*",
		"*.yay",
		"**.third",
		"nope.*",
	})
	if got != expect {
		t.Errorf("Expected %v, got %v", expect, got)
	}

	// Three-part missing
	expect, got = "", CombineActions([]string{
		"first.**",
		"hello",
		"*.second.*",
		"*.yay",
		"nope.*",
	})
	if got != expect {
		t.Errorf("Expected %v, got %v", expect, got)
	}
}
