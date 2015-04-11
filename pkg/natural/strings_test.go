// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import "testing"

func TestSplitQuoted(t *testing.T) {
	tests := map[string][]string{
		`a normal string`: {
			`a`,
			`normal`,
			`string`,
		},
		`a "quoted" string`: {
			`a`,
			`"quoted"`,
			`string`,
		},
		`a "longer quoted" string`: {
			`a`,
			`"longer quoted"`,
			`string`,
		},
		`a "multi quote" "string thing"`: {
			`a`,
			`"multi quote"`,
			`"string thing"`,
		},
		`a special"kind of""wrongness" here`: {
			`a`,
			`special"kind of""wrongness"`,
			`here`,
		},
		`oh, it's an "open quote`: {
			`oh,`,
			`it's`,
			`an`,
			`"open quote`,
		},
	}

	for s, exp := range tests {
		got, _ := SplitQuoted(s, " ")
		if len(got) != len(exp) {
			t.Errorf("wrong split: %#v", got)
			continue
		}
		for i := 0; i < len(got); i++ {
			if got[i] != exp[i] {
				t.Errorf("wrong split: %#v", got)
				break
			}
		}
	}
}

func TestTrimQuotes(t *testing.T) {
	tests := map[string]string{
		`unquoted text`: `unquoted text`,
		`"quoted text"`: `quoted text`,
		`"half quotes`:  `"half quotes`,
		`other half"`:   `other half"`,
	}
	for s, exp := range tests {
		got := TrimQuotes(s)
		if exp != got {
			t.Errorf("expected `%s`, got `%s`", exp, got)
		}
	}
}
