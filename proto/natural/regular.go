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
	Payload map[string]interface{}
}

type RegularSchemata []RegularSchema

func fillMessage(msg *proto.Message, payload map[string]interface{}, key, val string) {
	switch key {
	case "":
		return
	case "action":
		msg.Action = val
	case "text":
		msg.Text = val
	default:
		payload[key] = val
	}
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
		s.Payload = make(map[string]interface{})
		var err error
		re := `^(?i)` + regexp.QuoteMeta(s.Example) + `$`
		for field, val := range s.Fields {
			switch v := val.(type) {
			case string:
				fillMessage(&schemata[i].Message, s.Payload, field, v)
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
	payload := make(map[string]interface{})
	for k, v := range s.Payload {
		payload[k] = v
	}
	for i, field := range s.Regexp.SubexpNames() {
		fillMessage(&msg, payload, field, match[i])
	}
	if len(payload) > 0 {
		msg.EncodePayload(payload)
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
