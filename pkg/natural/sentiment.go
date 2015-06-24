// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

func AnalyzeAffirmativeSentiment(tokens []*Token) (string, float64) {
	inverse := false
	var pos, neg int

	for _, t := range tokens {
		tag := AffirmativeSentiment[t.Lemma]
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

var AffirmativeSentiment = map[string]string{
	"yes":         "pos",
	"yeah":        "pos",
	"ya":          "pos",
	"right":       "pos",
	"ok":          "pos",
	"okay":        "pos",
	"confirm":     "pos",
	"accept":      "pos",
	"affirmative": "pos",
	"sure":        "pos",
	"go":          "pos",
	"good":        "pos",

	"no":       "neg",
	"nah":      "neg",
	"nope":     "neg",
	"never":    "neg",
	"negative": "neg",
	"don't":    "neg",
	"off":      "neg",

	"not": "inv",

	"cancel": "cancel",
	"abort":  "cancel",
	"quit":   "cancel",
	"exit":   "cancel",
	"return": "cancel",
}
