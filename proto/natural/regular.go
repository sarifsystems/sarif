// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"regexp"
	"strings"

	"github.com/xconstruct/stark/proto"
	"gopkg.in/yaml.v1"
)

type RegularSchema struct {
	Example string
	Regexp  *regexp.Regexp
	Fields  map[string]interface{}
	Message proto.Message
}

type RegularSchemata []RegularSchema

func fillMessage(msg *proto.Message, key, val string) {
	if key == "action" {
		msg.Action = val
		return
	}
	if msg.Payload == nil {
		msg.Payload = make(map[string]interface{})
	}
	msg.Payload[key] = val
}

func buildRegexp(re, field string, val []interface{}) string {
	valStr := val[0].(string)
	old := regexp.QuoteMeta(valStr)
	repl := regexp.QuoteMeta(valStr)
	multiWords := strings.ContainsRune(old, ' ')
	if len(val) == 1 {
		if multiWords {
			repl = `.+`
		} else {
			repl = `[^\s]+`
		}
	}
	repl = `(?P<` + field + `>` + repl + `)`
	return strings.Replace(re, old, repl, -1)
}

func LoadRegularSchemata(text string) (RegularSchemata, error) {
	schemata := make(RegularSchemata, 0)
	if err := yaml.Unmarshal([]byte(text), &schemata); err != nil {
		return schemata, err
	}

	for i, s := range schemata {
		var err error
		re := `^(?i)` + regexp.QuoteMeta(s.Example) + `$`
		for field, val := range s.Fields {
			switch v := val.(type) {
			case string:
				fillMessage(&schemata[i].Message, field, v)
			case []interface{}:
				re = buildRegexp(re, field, v)
			}
		}
		schemata[i].Regexp, err = regexp.Compile(re)
		if err != nil {
			return schemata, err
		}
	}

	return schemata, nil
}

func (s RegularSchema) Parse(text string) (proto.Message, bool) {
	msg := s.Message.Copy()
	match := s.Regexp.FindStringSubmatch(text)
	if match == nil {
		return msg, false
	}
	for i, field := range s.Regexp.SubexpNames() {
		if field != "" {
			fillMessage(&msg, field, match[i])
		}
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
