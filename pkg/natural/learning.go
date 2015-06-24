// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"bytes"
	"regexp"
	"sort"
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

type sentenceRule struct {
	Rule     string
	Regexp   *regexp.Regexp
	Priority int
}

type byPriority []*sentenceRule

func (a byPriority) Len() int           { return len(a) }
func (a byPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPriority) Less(i, j int) bool { return a[i].Priority > a[j].Priority }

var specialBytes = []byte(`\.+*?()|[]{}^$`)

func quoteMeta(s string) string {
	b := make([]byte, 2*len(s))

	j := 0
	for i := 0; i < len(s); i++ {
		if bytes.IndexByte(specialBytes, s[i]) >= 0 {
			b[j] = '\\'
			j++
		}
		b[j] = s[i]
		j++
	}
	return string(b[0:j])
}

func newSentenceRule(s string) (r *sentenceRule, err error) {
	s = strings.ToLower(strings.TrimRight(s, ".?! "))
	r = &sentenceRule{s, nil, 0}
	s = quoteMeta(s)
	s = reMatchVars.ReplaceAllStringFunc(s, buildRegexp)
	s = `^(?i)` + s + `$`

	r.Regexp, err = regexp.Compile(s)
	if err != nil {
		return nil, err
	}

	r.Priority = len(s) - 10*r.Regexp.NumSubexp()
	return r, nil
}

func inStringSlice(s string, ss []string) bool {
	for _, a := range ss {
		if s == a {
			return true
		}
	}
	return false
}

func (r *sentenceRule) Parse(s string) []Var {
	match := r.Regexp.FindStringSubmatch(s)
	if match == nil {
		return nil
	}

	vars := make([]Var, 0)
	for i, field := range r.Regexp.SubexpNames() {
		if field == "" {
			continue
		}
		vars = append(vars, Var{field, match[i]})
	}
	return vars
}

type LearningParser struct {
	sentences []*sentenceRule
	messages  map[string]*MessageSchema

	perceptron *mlearning.Perceptron
}

func NewLearningParser() *LearningParser {
	return &LearningParser{
		make([]*sentenceRule, 0),
		make(map[string]*MessageSchema),

		mlearning.NewPerceptron(),
	}
}

func (p *LearningParser) LearnSentence(s string) {
	r, err := newSentenceRule(s)
	if err != nil {
		panic(err)
	}

	p.sentences = append(p.sentences, r)
	sort.Sort(byPriority(p.sentences))
}

type Meaning struct {
	Action     string
	Object     string
	ObjectType string

	Words     []string
	Variables []Var
}

type Var struct {
	Name  string
	Value string
}

func (m Meaning) Features() []mlearning.Feature {
	fs := make([]mlearning.Feature, 0)
	fs = append(fs, "bias")

	if m.Action != "" {
		fs = append(fs, mlearning.Feature("first="+m.Action))
	}
	if m.Object != "" {
		fs = append(fs, mlearning.Feature("second="+m.Object))
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
	for _, v := range m.Variables {
		fs = append(fs, mlearning.Feature("var="+v.Name))
	}
	return fs
}

func (p LearningParser) ParseSentence(s string) *Meaning {
	s = strings.TrimRight(s, ".?! ")
	for _, r := range p.sentences {
		if v := r.Parse(s); v != nil {
			return &Meaning{
				Words:     strings.Split(r.Rule, " "),
				Variables: v,
			}
		}
	}

	return nil
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
	for _, v := range m.Variables {
		switch v.Name {
		case "_action":
			msg.Action = v.Value
		case "text":
			msg.Text = v.Value
		case "to":
			fallthrough
		case "that":
			if msg.Text == "" {
				msg.Text = v.Value
			}
		default:
			if _, ok := r.Fields[v.Name]; ok {
				pl[v.Name] = v.Value
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

func (p *LearningParser) ReinforceSentence(sentence, action string) {
	m := p.ParseSentence(sentence)
	if m == nil {
		return
	}

	feats := m.Features()
	guess, _ := p.perceptron.Predict(feats)
	p.perceptron.Update(mlearning.Class(action), guess, feats)
}

func (p *LearningParser) FindMessageForMeaning(m *Meaning) (*MessageSchema, float64) {
	guess, w := p.perceptron.Predict(m.Features())

	schema := p.messages[string(guess)]
	return schema, float64(w)
}

func (p *LearningParser) Parse(text string) (proto.Message, float64) {
	msg := proto.Message{}

	m := p.ParseSentence(text)
	if m == nil {
		return msg, 0
	}

	r, w := p.FindMessageForMeaning(m)
	if r == nil {
		return msg, 0
	}
	msg = r.Apply(m)
	return msg, w
}

func (p *LearningParser) ParseWithAction(text, action string) (proto.Message, bool) {
	msg := proto.Message{}

	m := p.ParseSentence(text)
	if m == nil {
		return msg, false
	}

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
	rules := make([]string, len(p.sentences))
	for i, r := range p.sentences {
		rules[i] = r.Rule
	}
	schemata := make([]*MessageSchema, 0, len(p.messages))
	for _, s := range p.messages {
		schemata = append(schemata, s)
	}
	return &Model{
		rules,
		schemata,
		p.perceptron.Model,
	}
}

func (p *LearningParser) LoadModel(m *Model) error {
	rules := make([]*sentenceRule, len(m.Rules))
	for i, s := range m.Rules {
		r, err := newSentenceRule(s)
		if err != nil {
			return err
		}
		rules[i] = r
	}
	schemata := make(map[string]*MessageSchema)
	for _, s := range m.Schemata {
		schemata[s.Action] = s
	}

	p.sentences = rules
	p.messages = schemata
	p.perceptron.Model = m.Model
	return nil
}
