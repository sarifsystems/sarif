// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"regexp"
	"sort"
	"strings"

	"github.com/xconstruct/stark/proto"
)

type sentenceRule struct {
	Rule     string
	Regexp   *regexp.Regexp
	Priority int
}

type byPriority []*sentenceRule

func (a byPriority) Len() int           { return len(a) }
func (a byPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPriority) Less(i, j int) bool { return a[i].Priority > a[j].Priority }

func newSentenceRule(s string) (r *sentenceRule, err error) {
	r = &sentenceRule{s, nil, 0}
	s = regexp.QuoteMeta(s)
	s = reMatchVars.ReplaceAllStringFunc(s, buildRegexp)
	s = `^(?i)` + s + `$`

	r.Regexp, err = regexp.Compile(s)
	if err != nil {
		return nil, err
	}

	r.Priority = len(s) - 10*r.Regexp.NumSubexp()
	return r, nil
}

func (r *sentenceRule) Parse(s string) map[string]string {
	match := r.Regexp.FindStringSubmatch(s)
	if match == nil {
		return nil
	}

	vars := make(map[string]string)
	for i, field := range r.Regexp.SubexpNames() {
		if field == "" {
			continue
		}
		vars[field] = match[i]
	}
	return vars
}

type LearningParser struct {
	sentences []*sentenceRule
	messages  map[string][]*messageRule
}

func NewLearningParser() *LearningParser {
	return &LearningParser{
		make([]*sentenceRule, 0),
		make(map[string][]*messageRule),
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
	Keywords  map[string]struct{}
	Variables map[string]string
}

func (p LearningParser) ParseSentence(s string) *Meaning {
	for _, r := range p.sentences {
		if v := r.Parse(s); v != nil {
			k := make(map[string]struct{})
			for _, s := range strings.Split(r.Rule, " ") {
				k[s] = struct{}{}
			}
			for f := range v {
				k[f] = struct{}{}
			}
			return &Meaning{k, v}
		}
	}

	return nil
}

type messageRule struct {
	Action string
	Fields map[string]string
}

func (r *messageRule) Keywords() []string {
	ts := strings.Split(r.Action, "/")
	for field := range r.Fields {
		ts = append(ts, field)
	}
	return ts
}

func (p *LearningParser) LearnMessage(msg proto.Message) {
	r := &messageRule{
		msg.Action,
		make(map[string]string),
	}

	var fields map[string]interface{}
	msg.DecodePayload(&fields)
	for k := range fields {
		r.Fields[k] = "string"
	}

	for _, t := range r.Keywords() {
		p.messages[t] = append(p.messages[t], r)
	}
}

func (p *LearningParser) findMessageForMeaning(m *Meaning) *messageRule {
	candidates := make(map[*messageRule]int)
	for kw := range m.Keywords {
		if rs, ok := p.messages[kw]; ok {
			for _, r := range rs {
				candidates[r]++
			}
		}
	}
	if len(candidates) == 0 {
		return nil
	}

	var rMax *messageRule
	wMax := 0
	for r, w := range candidates {
		if w > wMax {
			rMax, wMax = r, w
		}
	}

	return rMax
}

func (p *LearningParser) Parse(text string) (proto.Message, bool) {
	msg := proto.Message{}

	m := p.ParseSentence(text)
	if m == nil {
		return msg, false
	}

	r := p.findMessageForMeaning(m)
	if r == nil {
		return msg, false
	}

	msg.Action = r.Action
	for k, v := range m.Variables {
		switch k {
		case "action":
			msg.Action = v
			delete(m.Variables, k)
		case "text":
			msg.Text = v
			delete(m.Variables, k)
		}
	}
	if len(m.Variables) > 0 {
		msg.EncodePayload(&m.Variables)
	}
	return msg, true
}
