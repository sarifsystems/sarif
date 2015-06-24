// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

func AnalyzeSentenceFunction(tokens []*Token) string {
	if len(tokens) == 0 {
		return ""
	}

	first := true
	for _, t := range tokens {
		if first {
			if inStringSlice(t.Lemma, Interrogatives) {
				return "interrogative"
			}
			if t.Is("!") || t.Lemma == "stark" {
				continue
			}
			first = false
			if t.Is("O") || t.Is("N") {
				return "declarative"
			}
		}
	}

	if first {
		return "exclamatory"
	}

	return "imperative"
}

var Interrogatives = []string{
	"which",
	"what",
	"who",
	"whom",
	"where",
	"when",
	"how",
	"why",
}
