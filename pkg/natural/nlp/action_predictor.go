// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package nlp

import (
	"fmt"

	"github.com/xconstruct/stark/pkg/mlearning"
)

var DefaultStopWords = []string{
	"a",
	"about",
	"an",
	"are",
	"as",
	"at",
	"be",
	"by",
	"for",
	"from",
	"how",
	"i",
	"in",
	"is",
	"it",
	"of",
	"on",
	"or",
	"that",
	"the",
	"this",
	"to",
	"was",
	"what",
	"when",
	"where",
	"who",
	"will",
	"with",
}

func inStringSlice(s string, ss []string) bool {
	for _, a := range ss {
		if s == a {
			return true
		}
	}
	return false
}

type ActionPredictor struct {
	Perceptron *mlearning.Perceptron
}

func NewActionPredictor() *ActionPredictor {
	return &ActionPredictor{
		mlearning.NewPerceptron(),
	}
}

type Meaning struct {
	Subject    string `json:"subject,omitempty"`
	Predicate  string `json:"predicate,omitempty"`
	Object     string `json:"object,omitempty"`
	ObjectType string `json:"object_type,omitempty"`

	Tokens []*Token `json:"-"`
	Vars   []*Var   `json:"vars"`
}

func (m Meaning) Features() []mlearning.Feature {
	fs := make([]mlearning.Feature, 0)
	fs = append(fs, "bias")

	// if m.Predicate != "" {
	// 	fs = append(fs, mlearning.Feature("predicate="+m.Predicate))
	// }
	// if m.Object != "" {
	// 	fs = append(fs, mlearning.Feature("object="+m.Object))
	// }
	for i, t := range m.Tokens {
		if i == 0 {
			fs = append(fs, mlearning.Feature("first="+t.Lemma))
		}
		if i == 1 {
			fs = append(fs, mlearning.Feature("second="+t.Lemma))
		}

		fs = append(fs, mlearning.Feature("word="+t.Lemma))
	}
	for _, v := range m.Vars {
		fs = append(fs, mlearning.Feature("var="+v.Name))
	}
	return fs
}

func (p *ActionPredictor) ReinforceMeaning(m *Meaning, action string) {
	feats := m.Features()
	guess, _ := p.Perceptron.Predict(feats)
	p.Perceptron.Update(mlearning.Class(action), guess, feats)
}

func (p *ActionPredictor) Predict(m *Meaning) (string, float64) {
	guess, w := p.Perceptron.Predict(m.Features())
	return string(guess), float64(w)
}

func (p *ActionPredictor) PredictAll(m *Meaning) []mlearning.Prediction {
	return p.Perceptron.PredictAll(m.Features())
}

type Model struct {
	Rules    []string
	Schemata []*MessageSchema
	*mlearning.Model
}

func (p *ActionPredictor) Train(iterations int, dataset DataSet, tokenizer *Tokenizer) {
	set := &mlearning.SimpleIterator{}
	for _, data := range dataset {
		tok := tokenizer.Tokenize(data.Sentence)
		set.FeatureSlice = append(set.FeatureSlice, &Meaning{
			Tokens: tok,
			Vars:   data.Vars,
		})
		set.ClassSlice = append(set.ClassSlice, mlearning.Class(data.Action))
	}

	for it := 0; it < iterations; it++ {
		set.Reset(true)
		c, n := p.Perceptron.Train(set)
		fmt.Printf("ActionPredictor iter %d: %d/%d=%.3f\n", it, c, n, float64(c)/float64(n)*100)
	}
}
