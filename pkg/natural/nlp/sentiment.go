// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package nlp

import "github.com/xconstruct/stark/pkg/datasets/en"

func AnalyzeAffirmativeSentiment(tokens []*Token) (string, float64) {
	inverse := false
	var pos, neg int

	for _, t := range tokens {
		tag := en.AffirmativeSentiment[t.Lemma]
		switch tag {
		case "cancel":
			return "cancel", 0
		case "pos":
			if inverse {
				neg++
			} else {
				pos++
			}
			inverse = false
		case "neg":
			if inverse {
				pos++
			} else {
				neg++
			}
			inverse = false
		case "inv":
			inverse = !inverse
		}
	}
	if inverse {
		neg++
	}

	w := float64(pos - neg)
	if w > 0 {
		return "pos", w
	} else if w < 0 {
		return "neg", w
	}
	return "", w
}
