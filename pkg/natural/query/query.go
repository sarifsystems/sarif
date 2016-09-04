// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package query

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sarifsystems/sarif/pkg/natural"
	"github.com/sarifsystems/sarif/pkg/natural/nlp"
)

type Parser struct {
	KnownObjects map[string]string
	tokenizer    *nlp.Tokenizer
}

func NewParser() *Parser {
	return &Parser{
		KnownObjects: make(map[string]string),
		tokenizer:    nlp.NewTokenizer(),
	}
}

var Projections = map[string]string{
	"select": "list",
	"which":  "list",
	"what":   "list",
	"who":    "list",
	"where":  "list",
	"show":   "list",
	"find":   "list",
	"list":   "list",
	"how":    "count",
	"many":   "count",
	"count":  "count",
}

var Operators = map[string]string{
	"=":  "=",
	"==": "==",
	"!=": "!=",
	">":  ">",
	"<":  "<",
	">=": ">=",
	"<=": "<=",
	"^":  "^",
	"$":  "$",

	"equal":    "==",
	"like":     "==",
	"not":      "!",
	"over":     ">=",
	"greater":  ">=",
	"after":    ">=",
	"lower":    "<",
	"before":   "<",
	"start":    "^",
	"starts":   "^",
	"starting": "^",
	"end":      "$",
	"ends":     "$",
	"ending":   "$",
}

var Modes = map[string]string{
	"where":   "select",
	"with":    "select",
	"sort":    "sort",
	"sorted":  "sort",
	"order":   "sort",
	"ordered": "sort",
	"group":   "group",
	"grouped": "group",
}

var Skipped = map[string]bool{
	"with": true,
	"than": true,
	"in":   true,
	"to":   true,
}

type Meaning struct {
	Predicate  string
	Object     string
	Selections map[string]interface{}
}

func (m Meaning) String() string {
	if m.Predicate == "" || m.Object == "" {
		return ""
	}

	s := m.Predicate + " " + m.Object
	if len(m.Selections) > 0 {
		s += " where "
		first := true
		for attr, val := range m.Selections {
			if !first {
				s += " and "
			}
			v := fmt.Sprintf("%v", val)
			if strings.Contains(v, " ") {
				v = `"` + v + `"`
			}
			s += attr + " " + v
			first = false
		}
	}
	return s
}

func invert(op string) string {
	switch op {
	case "==":
		return "!="
	case "!=":
		return "=="
	case ">":
		return "<="
	case "<":
		return ">="
	case ">=":
		return "<"
	case "<=":
		return ">"
	case "!":
		return ""
	}
	return op
}

func (p *Parser) Parse(ctx *natural.Context) (*natural.ParseResult, error) {
	if ctx == nil {
		ctx = &natural.Context{}
	}
	r := &natural.ParseResult{
		Text: ctx.Text,
	}
	if ctx.Text == "" {
		return r, nil
	}

	m := Meaning{
		Selections: make(map[string]interface{}),
	}
	var mode, attr, op string
	tokens := p.tokenizer.Tokenize(ctx.Text)
	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]

		// First part of sentence: find predicate and object
		if m.Predicate == "" {
			if p, ok := Projections[tok.Lemma]; ok {
				m.Predicate = p
			} else {
				return r, errors.New("Expected projection instead of '" + tok.Value + "'")
			}
		}
		if m.Object == "" {
			if p, ok := Projections[tok.Lemma]; ok {
				m.Predicate = p
			} else {
				m.Object = pluralize(tok.Lemma)
			}
			// TODO: handle limit
			// TODO: handle "last"/"top"
			continue
		}

		// Check for mode switches
		if md, ok := Modes[tok.Lemma]; ok {
			mode = md
			continue
		}
		if tok.Lemma == "by" && mode != "group" {
			mode = "sort"
			continue
		}

		// Skip stop words
		if ok := Skipped[tok.Lemma]; ok {
			continue
		}

		if tok.Lemma == "," || tok.Lemma == "and" {
			attr, op = "", ""
			continue
		}

		// Parse attribute/values
		// TODO: handle group/sort
		if o, ok := Operators[tok.Lemma]; ok {
			if op == "!" {
				op = invert(o)
			}
			op = o
			if attr == "" {
				return r, errors.New("Unexpected operator '" + op + "', expecting attribute")
			}
			continue
		}

		if attr == "" {
			attr = tok.Lemma
		} else {
			val := natural.ParseValue(tok.Value)
			if op != "" && op != "==" {
				attr = attr + " " + op
			}
			m.Selections[attr] = val
			attr = ""
		}
	}

	// Build intent from meaning
	it := natural.Intent{
		Type:      "query",
		Weight:    1,
		ExtraInfo: m,
	}
	if action, ok := p.KnownObjects[m.Object]; ok {
		it.Message.Action = action
		it.Message.EncodePayload(m.Selections)
	} else {
		it.Message.Action = "store/scan/" + m.Object
		it.Message.EncodePayload(map[string]interface{}{
			"filter": m.Selections,
		})
	}
	r.Text = m.String()
	r.Intents = append(r.Intents, &it)

	return r, nil
}

func (p *Parser) IsObjectKnown(object string) bool {
	_, ok := p.KnownObjects[pluralize(object)]
	return ok
}

func pluralize(object string) string {
	return strings.Trim(object, "s") + "s"
}

func (p *Parser) RegisterObject(object, action string) {
	p.KnownObjects[pluralize(object)] = action
}
