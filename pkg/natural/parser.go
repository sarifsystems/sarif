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

	"github.com/davecgh/go-spew/spew"
	"github.com/xconstruct/stark/pkg/datasets/commands"
	"github.com/xconstruct/stark/pkg/datasets/twpos"
	"github.com/xconstruct/stark/pkg/mlearning"
	"github.com/xconstruct/stark/proto"
)

type model struct {
	Schema map[string]*MessageSchema
	Rules  SentenceRuleSet
	Parser *mlearning.Model
	Pos    *mlearning.Model
	Var    *mlearning.Model
}

type Parser struct {
	schema       *MessageSchemaStore
	regular      *RegularParser
	tokenizer    *Tokenizer
	parser       *LearningParser
	pos          *PosTagger
	meaning      *MeaningParser
	varPredictor *VarPredictor
}

func NewParser() *Parser {
	return &Parser{
		NewMessageSchemaStore(),
		NewRegularParser(),
		NewTokenizer(),
		NewLearningParser(),
		NewPosTagger(),
		NewMeaningParser(),
		NewVarPredictor(),
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
		p.schema.Messages,
		p.regular.Rules(),
		p.parser.Perceptron.Model,
		p.pos.Perceptron.Model,
		p.varPredictor.Perceptron.Model,
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
			return p.TrainModel()
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
	p.schema.Messages = model.Schema
	if err := p.regular.Load(model.Rules); err != nil {
		return err
	}
	p.parser.Perceptron.Model = model.Parser
	p.pos.Perceptron.Model = model.Pos
	p.varPredictor.Perceptron.Model = model.Var
	return nil
}

func (p *Parser) TrainModel() error {
	set, err := ReadDataSet(strings.NewReader(commands.Data))
	if err != nil {
		return err
	}
	ss, err := LoadCoNLL(strings.NewReader(twpos.Data))
	if err != nil {
		return err
	}

	p.parser.Train(3, set)
	p.schema.AddDataSet(set)
	p.regular.Load(DefaultRules)
	p.pos.Train(5, ss)
	p.varPredictor.Train(10, set)

	return nil
}

type ParseResult struct {
	Text    string        `json:"text"`
	Type    string        `json:"type"`
	Weight  float64       `json:"weight"`
	Meaning *Meaning      `json:"meaning"`
	Message proto.Message `json:"msg"`
	Tokens  []*Token      `json:"tokens"`
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

	if msg, ok := ParseSimple(text); ok {
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
		r.Type, r.Weight = AnalyzeAffirmativeSentiment(r.Tokens)
		return r, nil
	}

	p.pos.Predict(r.Tokens)
	p.ResolvePronouns(r.Tokens, ctx)

	// TODO
	r.Type = AnalyzeSentenceFunction(r.Tokens)
	var err error
	switch r.Type {
	case "declarative":
		if r.Meaning, err = p.meaning.ParseDeclarative(r.Tokens); err != nil {
			return r, err
		}
		action := "event"
		if r.Meaning.Variables["fact"] != "" {
			action = "concept"
		}
		r.Message, err = p.InventMessageForMeaning(action, r.Meaning)
		r.Weight = 1 // TODO
		return r, err

	case "exclamatory":
	case "interrogative":
		if r.Meaning, err = p.meaning.ParseInterrogative(r.Tokens); err != nil {
			return r, err
		}
		r.Message, err = p.InventMessageForMeaning("concepts/query", r.Meaning)
		r.Weight = 1 // TODO

	case "imperative":
		if r.Meaning, err = p.meaning.ParseImperative(r.Tokens); err != nil {
			return r, err
		}
		action, w := p.parser.Predict(r.Meaning)
		r.Weight = w

		mschema := p.schema.Get(action)
		if mschema != nil {
			vars := p.varPredictor.PredictTokens(r.Tokens, mschema.Action)
			spew.Dump(vars)
			for _, v := range vars {
				r.Meaning.Variables[v.Name] = v.Value
			}
			r.Message = mschema.Apply(r.Meaning)
		}
		return r, err
	}

	return r, nil
}

func (p *Parser) ResolvePronouns(ts []*Token, ctx *Context) {
	for _, t := range ts {
		if t.Is("O") {
			switch t.Lemma {
			case "me":
				t.Lemma = ctx.Sender
			case "i":
				t.Lemma = ctx.Sender
			case "you":
				t.Lemma = ctx.Recipient
			}
		}
	}
}

func (p *Parser) LearnRule(text string, action string) error {
	return p.regular.Learn(text, action)
}

func (p *Parser) ReinforceSentence(text string, action string) error {
	r, err := p.Parse(text, &Context{})
	if err != nil {
		return err
	}
	p.parser.ReinforceMeaning(r.Meaning, action)
	return nil
}

func (p *Parser) LearnMessage(msg *proto.Message) {
	p.schema.AddMessage(msg)
}

func (p *Parser) InventMessageForMeaning(action string, m *Meaning) (proto.Message, error) {
	msg := proto.Message{}

	if action != "" {
		msg.Action += "/" + action
	}
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
