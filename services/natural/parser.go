// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"strings"

	"github.com/xconstruct/stark/pkg/mlearning"
	"github.com/xconstruct/stark/pkg/natural"
	"github.com/xconstruct/stark/pkg/natural/twpos"
	"github.com/xconstruct/stark/proto"
)

type model struct {
	Rules  natural.SentenceRuleSet
	Parser *natural.Model
	Pos    *mlearning.Model
}

type Parser struct {
	regular   *natural.RegularParser
	tokenizer *natural.Tokenizer
	parser    *natural.LearningParser
	pos       *natural.PosTagger
	meaning   *natural.MeaningParser
}

func NewParser() *Parser {
	return &Parser{
		natural.NewRegularParser(),
		natural.NewTokenizer(),
		natural.NewLearningParser(),
		natural.NewPosTagger(),
		natural.NewMeaningParser(),
	}
}

func (p *Parser) SaveModel(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	model := model{
		p.regular.Rules(),
		p.parser.Model(),
		p.pos.Perceptron.Model,
	}
	if err := json.NewEncoder(gz).Encode(model); err != nil {
		return err
	}
	return gz.Close()
}

func (p *Parser) LoadModel(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			p.regular.Load(natural.DefaultRules)
			ss, err := natural.LoadCoNLL(strings.NewReader(twpos.Data))
			if err != nil {
				return err
			}
			p.pos.Train(5, ss)
			return nil
		}
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(gz)

	model := model{}
	if err := dec.Decode(&model); err != nil {
		return err
	}
	if err := p.regular.Load(model.Rules); err != nil {
		return err
	}
	if err := p.parser.LoadModel(model.Parser); err != nil {
		return err
	}
	p.pos.Perceptron.Model = model.Pos
	return nil
}

type ParseResult struct {
	Text    string           `json:"text"`
	Type    string           `json:"type"`
	Weight  float64          `json:"weight"`
	Meaning *natural.Meaning `json:"meaning"`
	Message proto.Message    `json:"msg"`
	Tokens  []*natural.Token `json:"tokens"`
}

type Context struct {
	ExpectedReply string
	Sender        string
	Recipient     string
}

func (p *Parser) Parse(text string, ctx *Context) (*ParseResult, error) {
	if ctx == nil {
		ctx = &Context{}
	}
	r := &ParseResult{
		Text: text,
	}
	if text == "" {
		return r, nil
	}

	if msg, ok := natural.ParseSimple(text); ok {
		r.Type = "simple"
		r.Message = msg
		r.Weight = 100
		return r, nil
	}

	if msg, ok := p.regular.Parse(text); ok {
		r.Type = "regular"
		r.Message = msg
		r.Weight = 100
		return r, nil
	}

	r.Tokens = p.tokenizer.Tokenize(text)

	if ctx.ExpectedReply == "affirmative" {
		r.Type, r.Weight = natural.AnalyzeAffirmativeSentiment(r.Tokens)
		return r, nil
	}

	p.pos.Predict(r.Tokens)
	p.ResolvePronouns(r.Tokens, ctx)

	// TODO
	r.Type = natural.AnalyzeSentenceFunction(r.Tokens)
	var err error
	switch r.Type {
	case "declarative":
		if r.Meaning, err = p.meaning.ParseDeclarative(r.Tokens); err != nil {
			return r, err
		}
		r.Message, err = p.InventMessageForMeaning(r.Meaning)
		r.Weight = 1 // TODO
		return r, err
	case "exclamatory":
	case "interrogative":
	case "imperative":
		if r.Meaning, err = p.meaning.ParseImperative(r.Tokens); err != nil {
			return r, err
		}
		mschema, w := p.parser.FindMessageForMeaning(r.Meaning)
		r.Weight = w
		if mschema != nil {
			r.Message = mschema.Apply(r.Meaning)
		}
		return r, err
	}

	return r, nil
}

func (p *Parser) ResolvePronouns(ts []*natural.Token, ctx *Context) {
	for _, t := range ts {
		if t.Is("O") {
			switch t.Lemma {
			case "i":
				t.Lemma = ctx.Sender
			case "you":
				t.Lemma = ctx.Recipient
			}
		}
	}
}

func (p *Parser) ReinforceSentence(text string, action string) error {
	r, err := p.Parse(text, &Context{})
	if err != nil {
		return err
	}
	p.parser.ReinforceMeaning(r.Meaning, action)
	return nil
}

func (p *Parser) InventMessageForMeaning(m *natural.Meaning) (proto.Message, error) {
	msg := proto.Message{}

	if m.Subject != "" {
		msg.Action += "/" + m.Subject
	}
	if m.Predicate != "" {
		msg.Action += "/" + m.Predicate
	}
	if m.Object != "" {
		msg.Action += "/" + m.Object
	}

	msg.Action = strings.Trim(msg.Action, "/")
	msg.EncodePayload(m.Variables)
	return msg, nil
}

func inStringSlice(s string, ss []string) bool {
	for _, a := range ss {
		if s == a {
			return true
		}
	}
	return false
}
