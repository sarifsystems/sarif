// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"strings"
	"testing"
)

type CallingTest struct {
	Action string
	CountA int
	CountB int
}

func TestSubtree(t *testing.T) {
	a, b := NewPipe()
	st := newSubtree()

	st.Subscribe(strings.Split("basic/test/topic", "/"), a)
	st.Subscribe(strings.Split("basic/test/topic", "/"), b)
	st.Subscribe(strings.Split("basic/other/topic", "/"), b)

	st.Subscribe(strings.Split("multi/topic", "/"), a)
	st.Subscribe(strings.Split("multi", "/"), a)

	st.Subscribe(strings.Split("unsub/some/topic", "/"), a)
	st.Unsubscribe(strings.Split("unsub", "/"), a)

	st.Subscribe(strings.Split("widening/topic/sub", "/"), a)
	st.Subscribe(strings.Split("widening/topic/sub", "/"), b)
	st.Subscribe(strings.Split("widening/topic", "/"), a)
	st.Subscribe(strings.Split("widening/topic/another", "/"), a)

	aFired, bFired := 0, 0
	count := func(c writer) {
		if c == a {
			aFired++
		} else if c == b {
			bFired++
		}
	}

	tests := []CallingTest{
		{"", 0, 0},
		{"basic", 0, 0},
		{"basic/test/topic", 1, 1},
		{"basic/other/topic", 0, 1},
		{"basic/no/topic", 0, 0},

		{"multi/topic", 1, 0},

		{"unsub/some/topic", 0, 0},

		{"widening/topic/sub", 1, 1},
		{"widening/topic", 1, 0},
		{"widening/topic/another", 1, 0},
	}

	for i, test := range tests {
		action := strings.Split(test.Action, "/")
		aFired, bFired = 0, 0
		st.Call(action, count)

		t.Logf("test %d: %s", i, test.Action)
		if aFired != test.CountA {
			t.Errorf("    expected %d calls of A, got %d", test.CountA, aFired)
		}
		if bFired != test.CountB {
			t.Errorf("    expected %d calls of B, got %d", test.CountB, bFired)
		}
	}
}
