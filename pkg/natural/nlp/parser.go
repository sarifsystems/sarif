// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package nlp

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"strings"

	"github.com/sarifsystems/sarif/pkg/datasets/commands"
	"github.com/sarifsystems/sarif/pkg/datasets/twpos"
	"github.com/sarifsystems/sarif/pkg/mlearning"
	"github.com/sarifsystems/sarif/pkg/natural"
	"github.com/sarifsystems/sarif/sarif"
)

type model struct {
	Schema map[string]*MessageSchema
	Parser *mlearning.Model
	Pos    *mlearning.Model
	Var    *mlearning.Model
}

type Parser struct {
	schema       *MessageSchemaStore
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
		tok,
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
	p.pos.Train(5, ss)
	p.varPredictor.Train(10, set, p.tokenizer)

	return nil
}

type ParseResult struct {
	Text    string            `json:"text"`
	Intents []*natural.Intent `json:"intents"`

	Meaning *Meaning `json:"meaning"`
	Tokens  []*Token `json:"tokens"`
}

func (r ParseResult) String() string {
	s := "Interpretation: " + r.Text
	if len(r.Intents) > 0 {
		s += "\n"
		for _, intent := range r.Intents {
			s += "\n" + intent.String()
		}
	} else if r.Meaning != nil {
		for _, v := range r.Meaning.Vars {
			s += " " + v.String()
		}
	}
	return s
}

func (p *Parser) Parse(ctx *natural.Context) (*ParseResult, error) {
	if ctx == nil {
		ctx = &natural.Context{}
	}
	r := &ParseResult{
		Text: ctx.Text,
	}
	if ctx.Text == "" {
		return r, nil
	}

	r.Tokens = p.tokenizer.Tokenize(ctx.Text)

	if ctx.ExpectedReply == "affirmative" {
		typ, w := AnalyzeAffirmativeSentiment(r.Tokens)
		r.Intents = []*natural.Intent{{Type: typ, Weight: w}}
		return r, nil
	}

	p.pos.Predict(r.Tokens)
	p.ResolvePronouns(r.Tokens, ctx)

	typ := AnalyzeSentenceFunction(r.Tokens)
	var msg sarif.Message
	var err error
	switch typ {
	case "declarative":
		if r.Meaning, err = p.meaning.ParseDeclarative(r.Tokens); err != nil {
			return r, err
		}
		action := "event"
		msg, err = p.InventMessageForMeaning(action, r.Meaning)
		r.Intents = []*natural.Intent{{
			Intent: msg.Action,
			Type:   typ,
			Weight: 0.25, // TODO

			Message:   msg,
			ExtraInfo: r.Meaning.Vars,
		}}

	case "exclamatory":
		r.Intents = []*natural.Intent{{
			Type:   typ,
			Weight: 0.5,
		}}
	case "interrogative":
		if r.Meaning, err = p.meaning.ParseInterrogative(r.Tokens); err != nil {
			return r, err
		}
		msg, err = p.InventMessageForMeaning("concepts/query", r.Meaning)
		r.Intents = []*natural.Intent{{
			Intent: msg.Action,
			Type:   typ,
			Weight: 0.25, // TODO

			Message:   msg,
			ExtraInfo: r.Meaning.Vars,
		}}

	case "imperative":
		if r.Meaning, err = p.meaning.ParseImperative(r.Tokens); err != nil {
			return r, err
		}

		preds := p.action.PredictAll(r.Meaning)
		for i, pred := range preds {
			if i > 4 {
				break
			}

			fp := &natural.Intent{
				Intent: string(pred.Class),
				Type:   typ,
				Weight: float64(pred.Weight) / 10, // TODO: Weights
			}
			r.Intents = append(r.Intents, fp)

			if mschema := p.schema.Get(fp.Intent); mschema != nil {
				vars := p.varPredictor.PredictTokens(r.Tokens, mschema.Action)
				vars = append(vars, r.Meaning.Vars...)
				fp.Message = mschema.Apply(vars)
				fp.ExtraInfo = vars
			}
		}

		return r, err
	}

	return r, nil
}

func (p *Parser) ResolvePronouns(ts []*Token, ctx *natural.Context) {
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

func (p *Parser) ReinforceSentence(text string, action string) error {
	r, err := p.Parse(&natural.Context{Text: text})
	if err != nil {
		return err
	}
	p.action.ReinforceMeaning(r.Meaning, action)
	return nil
}

func (p *Parser) LearnMessage(msg *sarif.Message) {
	p.schema.AddMessage(msg)
}

func (p *Parser) InventMessageForMeaning(action string, m *Meaning) (sarif.Message, error) {
	msg := sarif.Message{}

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
