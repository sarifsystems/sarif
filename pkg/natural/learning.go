// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"strings"

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
	messages map[string]*MessageSchema

	perceptron *mlearning.Perceptron
}

func NewLearningParser() *LearningParser {
	return &LearningParser{
		make(map[string]*MessageSchema),

		mlearning.NewPerceptron(),
	}
}

type Meaning struct {
	Subject    string `json:"subject,omitempty"`
	Predicate  string `json:"predicate,omitempty"`
	Object     string `json:"object,omitempty"`
	ObjectType string `json:"object_type,omitempty"`

	Words     []string
	Variables map[string]string `json:"variables"`
}

type Var struct {
	Name  string
	Value string
}

func (m Meaning) Features() []mlearning.Feature {
	fs := make([]mlearning.Feature, 0)
	fs = append(fs, "bias")

	if m.Predicate != "" {
		fs = append(fs, mlearning.Feature("predicate="+m.Predicate))
	}
	if m.Object != "" {
		fs = append(fs, mlearning.Feature("object="+m.Object))
	}
	for i, w := range m.Words {
		if w[0] == '[' {
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
	return fs
}

type MessageSchema struct {
	Action string
	Fields map[string]string
}

func (r *MessageSchema) Features() []mlearning.Feature {
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

func (r *MessageSchema) Apply(m *Meaning) proto.Message {
	msg := proto.Message{}
	msg.Action = r.Action
	pl := make(map[string]string)
	for name, value := range m.Variables {
		switch name {
		case "_action":
			msg.Action = value
		case "text":
			msg.Text = value
		case "to":
			fallthrough
		case "that":
			if msg.Text == "" {
				msg.Text = value
			}
		default:
			if _, ok := r.Fields[name]; ok {
				pl[name] = value
			}
		}
	}
	if len(m.Variables) > 0 {
		msg.EncodePayload(&pl)
	}
	return msg
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
	if _, ok := p.messages[s.Action]; ok {
		// TODO: Merge fields
		return
	}
	p.messages[s.Action] = &s

	feats := s.Features()
	guess, _ := p.perceptron.Predict(feats)
	p.perceptron.Update(mlearning.Class(s.Action), guess, feats)
}

func (p *LearningParser) ReinforceMeaning(m *Meaning, action string) {
	feats := m.Features()
	guess, _ := p.perceptron.Predict(feats)
	p.perceptron.Update(mlearning.Class(action), guess, feats)
}

func (p *LearningParser) FindMessageForMeaning(m *Meaning) (*MessageSchema, float64) {
	guess, w := p.perceptron.Predict(m.Features())

	schema := p.messages[string(guess)]
	return schema, float64(w)
}

func (p *LearningParser) Parse(m *Meaning) (proto.Message, float64) {
	msg := proto.Message{}

	r, w := p.FindMessageForMeaning(m)
	if r == nil {
		return msg, 0
	}
	msg = r.Apply(m)
	return msg, w
}

func (p *LearningParser) ParseWithAction(m *Meaning, action string) (proto.Message, bool) {
	msg := proto.Message{}

	s, ok := p.messages[action]
	if !ok {
		return msg, false
	}

	return s.Apply(m), true
}

type Model struct {
	Rules    []string
	Schemata []*MessageSchema
	*mlearning.Model
}

func (p *LearningParser) Model() *Model {
	schemata := make([]*MessageSchema, 0, len(p.messages))
	for _, s := range p.messages {
		schemata = append(schemata, s)
	}
	return &Model{
		nil,
		schemata,
		p.perceptron.Model,
	}
}

func (p *LearningParser) LoadModel(m *Model) error {
	rules := make([]*SentenceRule, len(m.Rules))
	for i, s := range m.Rules {
		r, err := CompileSentenceRule(s, "")
		if err != nil {
			return err
		}
		rules[i] = r
	}
	schemata := make(map[string]*MessageSchema)
	for _, s := range m.Schemata {
		schemata[s.Action] = s
	}

	p.messages = schemata
	p.perceptron.Model = m.Model
	return nil
}
