// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package query

import (
	"testing"

	"github.com/sarifsystems/sarif/pkg/natural"
)

type Test struct {
	Input   string
	Meaning string
}

func TestParsing(t *testing.T) {
	tests := []Test{
		{
			`list events where time >= 5pm`,
			`list events where time >= 5pm`,
		},
		{
			`find contact with age lower than 40`,
			`list contacts where and age < 40`,
		},
		{
			`show location with address like Berlin`,
			`list locations where address Berlin`,
		},
		{
			`count events where action starts with timetracker`,
			`count events where action ^ timetracker`,
		},
		{
			`events where action like browser`,
			`error`,
		},
	}
	p := NewParser()

	for _, test := range tests {
		r, err := p.Parse(&natural.Context{Text: test.Input})
		if test.Meaning == "error" {
			if err == nil {
				t.Errorf("Expected error, but got:\n%q", r.Text)
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if r.Text != test.Meaning {
			t.Errorf("\nExpected: %q\nGot     : %q", test.Meaning, r.Text)
		}
	}
}
