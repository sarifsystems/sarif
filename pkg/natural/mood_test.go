// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import "testing"

func TestAnalyzeSentenceFunction(t *testing.T) {
	tests := map[string]string{
		"remind:V me:O to:P make:V coffee:N":             "imperative",
		"play:V some:D music:N":                          "imperative",
		"stark:N do:V some:D magic:N":                    "imperative",
		"when:O did:V i:O last:A make:V coffee:N ?:,":    "interrogative",
		"what:O is:V the:D capital:N of:P Germany:N ?:,": "interrogative",
		"where:O are:V you:O":                            "interrogative",
		"oh:! my:! god:!":                                "exclamatory",
		"awesome:!":                                      "exclamatory",
		"things:N are:V good:A":                          "declarative",
		"alice:N is:V tired:A":                           "declarative",

		"unknown:A": "imperative",
	}
	for sentence, exp := range tests {
		got := AnalyzeSentenceFunction(tagged(sentence))
		if got != exp {
			t.Errorf("sentence %q: got %q, expected %q", sentence, got, exp)
		}
	}
}
