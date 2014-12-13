// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/xconstruct/stark/pkg/template"
	"github.com/xconstruct/stark/proto"
)

var reMatchVars = regexp.MustCompile(`\\\[([\w ]+)\\\]`)

type RegularSchema struct {
	Example string           `json:"example"`
	Message *json.RawMessage `json:"msg"`

	Regexp   *regexp.Regexp     `json:"-"`
	Template *template.Template `json:"-"`
}

type RegularSchemata []RegularSchema

func buildRegexp(field string) string {
	matches := reMatchVars.FindStringSubmatch(field)
	field = matches[1]
	repl := `[^\s]+`
	if strings.ContainsRune(field, ' ') {
		repl = `.+`
	}
	field = strings.Replace(field, " ", "", -1)
	return `(?P<` + field + `>` + repl + `)`
}

func LoadRegularSchemata(text string) (RegularSchemata, error) {
	schemata := make(RegularSchemata, 0)
	if err := json.Unmarshal([]byte(text), &schemata); err != nil {
		return schemata, err
	}

	var err error
	for i, s := range schemata {
		re := regexp.QuoteMeta(s.Example)
		re = reMatchVars.ReplaceAllStringFunc(re, buildRegexp)
		re = `^(?i)` + re + `$`
		schemata[i].Regexp, err = regexp.Compile(re)
		if err != nil {
			return schemata, err
		}

		schemata[i].Template, err = template.New().Parse(string(*s.Message))
		if err != nil {
			return schemata, err
		}
	}

	return schemata, nil
}

func (s RegularSchema) Parse(text string) (proto.Message, bool) {
	match := s.Regexp.FindStringSubmatch(text)
	if match == nil {
		return proto.Message{}, false
	}

	vars := make(map[string]string)
	for i, field := range s.Regexp.SubexpNames() {
		vars[field] = match[i]
	}
	var msg proto.Message
	if err := s.Template.Execute(&msg, vars); err != nil {
		panic(err)
	}
	if msg.Text == "" {
		msg.Text = text
	}
	return msg, true
}

func (st RegularSchemata) Parse(text string) (proto.Message, bool) {
	text = strings.TrimRight(text, ".!?")
	for _, s := range st {
		if msg, ok := s.Parse(text); ok {
			return msg, true
		}
	}
	return proto.Message{}, false
}
