// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import "testing"

func TestAnalyzeAffirmativeSentiment(t *testing.T) {
	tests := map[string]string{
		"yes":   "pos",
		"nope":  "neg",
		"abort": "cancel",

		"no way":           "neg",
		"certainly not":    "neg",
		"do not do this":   "neg",
		"not right":        "neg",
		"sure":             "pos",
		"not sure":         "neg",
		"i did not say no": "pos",

		"yeah no": "",
	}
	for sentence, exp := range tests {
		got, _ := AnalyzeAffirmativeSentiment(tagged(sentence))
		if got != exp {
			t.Errorf("sentence %q: got %q, expected %q", sentence, got, exp)
		}
	}
}
