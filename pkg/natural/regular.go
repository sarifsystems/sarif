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

	"github.com/sarifsystems/sarif/sarif"
)

var reMatchVars = regexp.MustCompile(`\\\[([\w ]+)\\\]`)

func buildRegexp(field string) string {
	matches := reMatchVars.FindStringSubmatch(field)
	field = matches[1]
	field = strings.Replace(field, " ", "", -1)
	return `(?P<` + field + `>("[^"]*"|[^"].*))`
}

type RegularParser struct {
	rules []*SentenceRule
}

func NewRegularParser() *RegularParser {
	return &RegularParser{
		make([]*SentenceRule, 0),
	}
}

func (p *RegularParser) Parse(s string) (sarif.Message, bool) {
	s = strings.TrimRight(s, ".?! ")
	for _, r := range p.rules {
		if msg, ok := r.Parse(s); ok {
			return msg, ok
		}
	}
	return sarif.Message{}, false
}

func (p *RegularParser) Learn(rule, action string) error {
	r, err := CompileSentenceRule(rule, action)
	if err != nil {
		return err
	}
	p.rules = append(p.rules, r)
	sort.Sort(byPriority(p.rules))
	return nil
}

type SentenceRuleSet map[string]string

func (p *RegularParser) Rules() SentenceRuleSet {
	s := SentenceRuleSet{}
	for _, r := range p.rules {
		s[r.Rule] = r.Action
	}
	return s
}

func (p *RegularParser) Load(s SentenceRuleSet) error {
	for rule, action := range s {
		if err := p.Learn(rule, action); err != nil {
			return err
		}
	}
	return nil
}

type SentenceRule struct {
	Rule     string
	Action   string
	Regexp   *regexp.Regexp
	Priority int
}

type byPriority []*SentenceRule

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

func CompileSentenceRule(s, action string) (r *SentenceRule, err error) {
	s = strings.ToLower(strings.TrimRight(s, ".?! "))
	r = &SentenceRule{s, action, nil, 0}
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

func (r *SentenceRule) Parse(s string) (sarif.Message, bool) {
	msg := sarif.Message{}
	match := r.Regexp.FindStringSubmatch(s)
	if match == nil {
		return msg, false
	}

	msg.Action = r.Action
	p := make(map[string]string)
	for i, field := range r.Regexp.SubexpNames() {
		if field == "" {
			continue
		}
		v := TrimQuotes(match[i])
		if field == "text" {
			msg.Text = v
			continue
		}
		p[field] = v
	}
	if len(p) > 0 {
		msg.EncodePayload(p)
	}
	return msg, true
}
