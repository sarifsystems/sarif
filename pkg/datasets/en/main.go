// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package en contains a dataset of english words and phrases.
package en

var Contractions = map[string]string{
	"you're": "you are",
	"I'm":    "I am",
	"I've":   "I have",
	"you've": "you have",
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
