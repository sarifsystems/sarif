// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"strings"

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
	action       *ActionPredictor
	pos          *PosTagger
	meaning      *MeaningParser
	varPredictor *VarPredictor
}

func NewParser() *Parser {
	tok := NewTokenizer()
	return &Parser{
		NewMessageSchemaStore(),
		NewRegularParser(),
		NewTokenizer(),
		NewActionPredictor(),
		NewPosTagger(),
		NewMeaningParser(),
		NewVarPredictor(tok),
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
		p.action.Perceptron.Model,
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
	p.action.Perceptron.Model = model.Parser
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

	p.action.Train(10, set, p.tokenizer)
	p.schema.AddDataSet(set)
	p.regular.Load(DefaultRules)
	p.pos.Train(5, ss)
	p.varPredictor.Train(10, set, p.tokenizer)

	return nil
}

type ParseResult struct {
	Text    string   `json:"text"`
	Type    string   `json:"type"`
	Meaning *Meaning `json:"meaning"`
	Tokens  []*Token `json:"tokens"`

	Prediction  *Prediction
	Predictions []*Prediction `json:"predictions"`
}

func (r ParseResult) String() string {
	s := "Type: " + r.Type
	s += "\nInterpretation: " + r.Text
	if r.Prediction != nil {
		for _, v := range r.Prediction.Vars {
			s += " " + v.String()
		}
		s += "\n\nIntent: " + r.Prediction.String()
	} else if r.Meaning != nil {
		for _, v := range r.Meaning.Vars {
			s += " " + v.String()
		}
	}
	return s
}

type Context struct {
	ExpectedReply string
	Sender        string
	Recipient     string
}

type Prediction struct {
	Action  string        `json:"action"`
	Vars    []*Var        `json:"vars"`
	Message proto.Message `json:"msg"`
	Weight  float64       `json:"weight"`
}

func (p Prediction) String() string {
	s := "." + p.Message.Action
	v := make(map[string]string)
	p.Message.DecodePayload(&v)
	for name, val := range v {
		s += " " + name + "=" + val
	}
	s += fmt.Sprintf(" [weight: %g]", p.Weight)
	return s
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
		r.Prediction = &Prediction{
			Action:  msg.Action,
			Message: msg,
			Weight:  100,
		}
		return r, nil
	}

	if msg, ok := p.regular.Parse(text); ok {
		r.Type = "regular"
		r.Prediction = &Prediction{
			Action:  msg.Action,
			Message: msg,
			Weight:  100,
		}
		return r, nil
	}

	r.Tokens = p.tokenizer.Tokenize(text)

	if ctx.ExpectedReply == "affirmative" {
		var w float64
		r.Type, w = AnalyzeAffirmativeSentiment(r.Tokens)
		r.Prediction = &Prediction{Weight: w}
		return r, nil
	}

	p.pos.Predict(r.Tokens)
	p.ResolvePronouns(r.Tokens, ctx)

	// TODO
	r.Type = AnalyzeSentenceFunction(r.Tokens)
	var msg proto.Message
	var err error
	switch r.Type {
	case "declarative":
		if r.Meaning, err = p.meaning.ParseDeclarative(r.Tokens); err != nil {
			return r, err
		}
		action := "event"
		// if r.Meaning.Vars["fact"] != "" {
		// 	action = "concept"
		// }
		msg, err = p.InventMessageForMeaning(action, r.Meaning)
		r.Prediction = &Prediction{
			Action:  msg.Action,
			Message: msg,
			Vars:    r.Meaning.Vars,
			Weight:  1, // TODO
		}

	case "exclamatory":
	case "interrogative":
		if r.Meaning, err = p.meaning.ParseInterrogative(r.Tokens); err != nil {
			return r, err
		}
		msg, err = p.InventMessageForMeaning("concepts/query", r.Meaning)
		r.Prediction = &Prediction{
			Action:  msg.Action,
			Message: msg,
			Vars:    r.Meaning.Vars,
			Weight:  1, // TODO
		}

	case "imperative":
		if r.Meaning, err = p.meaning.ParseImperative(r.Tokens); err != nil {
			return r, err
		}

		preds := p.action.PredictAll(r.Meaning)
		first := true
		for i, pred := range preds {
			if i > 9 && !first {
				break
			}

			fp := &Prediction{
				Action: string(pred.Class),
				Weight: float64(pred.Weight),
				Vars:   r.Meaning.Vars,
			}
			r.Predictions = append(r.Predictions, fp)
			if r.Prediction != nil && (r.Prediction.Weight-fp.Weight) < 1 {
				r.Prediction = nil
			}

			if mschema := p.schema.Get(fp.Action); mschema != nil {
				vars := p.varPredictor.PredictTokens(r.Tokens, mschema.Action)
				fp.Vars = append(fp.Vars, vars...)
				fp.Message = mschema.Apply(fp.Vars)

				if first {
					first = false
					r.Prediction = fp
				}
			}
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
	p.action.ReinforceMeaning(r.Meaning, action)
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
	pl := make(map[string]string)
	for _, v := range m.Vars {
		pl[v.Name] = v.Value
	}
	msg.EncodePayload(pl)
	return msg, nil
}
