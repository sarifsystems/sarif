// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package nlp

import "github.com/sarifsystems/sarif/pkg/datasets/en"

func AnalyzeSentenceFunction(tokens []*Token) string {
	if len(tokens) == 0 {
		return ""
	}

	first := true
	var subject, predicate bool
	for _, t := range tokens {
		if first {
			if inStringSlice(t.Lemma, en.Interrogatives) {
				return "interrogative"
			}
			if t.Is("!") || t.Is("D") {
				continue
			}
			first = false
		}

		if !subject && !predicate && (t.Is("O") || t.Is("N")) {
			subject = true
		}
		if t.Is("V") {
			predicate = true
		}
	}

	if first {
		return "exclamatory"
	}
	if subject && predicate {
		return "declarative"
	}
	if predicate {
		return "imperative"
	}
	if len(tokens) <= 3 {
		return "exclamatory"
	}

	return "imperative"
}
