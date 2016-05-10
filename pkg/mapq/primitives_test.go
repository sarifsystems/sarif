// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mapq

import "testing"

type MatchTest struct {
	A      interface{}
	Op     string
	B      interface{}
	Result bool
}

func TestSimpleMatches(t *testing.T) {
	tests := []MatchTest{
		{"hello", "==", "hello", true},
		{"abc", "<", "xyz", true},
		{"prefixstring", "^", "prefix", true},
		{"prefixstring", "^", "notprefix", false},

		{5, ">", 3, true},
		{4, "==", 4, true},
		{4.3, "<", 5, true},

		{"0", "==", 0, false},
		{"0", "!=", 0, false},
		{0.0, "==", 0, true},
	}
	for i, test := range tests {
		if Matches(test.A, test.Op, test.B) != test.Result {
			t.Errorf(`Test %d failed: "%v" %s "%v" is %v`,
				i, test.A, test.Op, test.B, !test.Result)
		}
	}
}
