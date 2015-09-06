// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"fmt"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/xconstruct/stark/pkg/mlearning"
	"github.com/xconstruct/stark/proto"
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

type LearningParser struct {
	Perceptron *mlearning.Perceptron
}

func NewLearningParser() *LearningParser {
	return &LearningParser{
		mlearning.NewPerceptron(),
	}
}

type Meaning struct {
	Subject    string `json:"subject,omitempty"`
	Predicate  string `json:"predicate,omitempty"`
	Object     string `json:"object,omitempty"`
	ObjectType string `json:"object_type,omitempty"`

	Words     []string          `json:"words"`
	Variables map[string]string `json:"variables"`
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
	for i, w := range m.Words {
		if w == "" || w[0] == '[' {
			continue
		}
		if i == 0 {
			fs = append(fs, mlearning.Feature("first="+w))
		}
		if i == 1 {
			fs = append(fs, mlearning.Feature("second="+w))
		}

		if !inStringSlice(w, DefaultStopWords) {
			fs = append(fs, mlearning.Feature("word="+w))
		}
	}
	for name, _ := range m.Variables {
		fs = append(fs, mlearning.Feature("var="+name))
	}
	spew.Dump(fs)
	return fs
}

func getMessageSchemaFeatures(r MessageSchema) []mlearning.Feature {
	fs := make([]mlearning.Feature, 0)
	fs = append(fs, "bias")
	for _, p := range strings.Split(r.Action, "/") {
		fs = append(fs, mlearning.Feature("word="+p))
	}
	for field := range r.Fields {
		fs = append(fs, mlearning.Feature("var="+field))
	}
	return fs
}

func (p *LearningParser) LearnMessage(msg proto.Message) {
	s := MessageSchema{
		msg.Action,
		make(map[string]string),
	}
	var fields map[string]interface{}
	msg.DecodePayload(&fields)
	for k := range fields {
		s.Fields[k] = "string"
	}

	p.LearnMessageSchema(s)
}

func (p *LearningParser) LearnMessageSchema(s MessageSchema) {
	if s.Action == "" {
		return
	}
	feats := getMessageSchemaFeatures(s)
	guess, _ := p.Perceptron.Predict(feats)
	p.Perceptron.Update(mlearning.Class(s.Action), guess, feats)
}

func (p *LearningParser) ReinforceMeaning(m *Meaning, action string) {
	feats := m.Features()
	guess, _ := p.Perceptron.Predict(feats)
	p.Perceptron.Update(mlearning.Class(action), guess, feats)
}

func (p *LearningParser) Predict(m *Meaning) (string, float64) {
	guess, w := p.Perceptron.Predict(m.Features())
	return string(guess), float64(w)
}

type Model struct {
	Rules    []string
	Schemata []*MessageSchema
	*mlearning.Model
}

func (p *LearningParser) Train(iterations int, dataset DataSet) {
	set := &mlearning.SimpleIterator{}
	for _, data := range dataset {
		vars := make(map[string]string)
		for _, v := range data.Variables {
			vars[v.Name] = v.Type
		}
		set.FeatureSlice = append(set.FeatureSlice, &Meaning{
			Words:     strings.Split(data.CleanedSentence(""), " "), // TODO: tokenize
			Variables: vars,
		})
		set.ClassSlice = append(set.ClassSlice, mlearning.Class(data.Action))
	}

	for it := 0; it < iterations; it++ {
		set.Reset(true)
		c, n := p.Perceptron.Train(set)
		fmt.Printf("LearningParser iter %d: %d/%d=%.3f\n", it, c, n, float64(c)/float64(n)*100)
	}
}
