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
	"sort"
	"strings"

	"github.com/xconstruct/stark/pkg/datasets/commands"
	"github.com/xconstruct/stark/pkg/datasets/twpos"
	"github.com/xconstruct/stark/pkg/mlearning"
	"github.com/xconstruct/stark/proto"
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
	Text    string   `json:"text"`
	Meaning *Meaning `json:"meaning"`
	Tokens  []*Token `json:"tokens"`

	Predictions []*Prediction `json:"predictions"`
}

func (r ParseResult) String() string {
	s := "Interpretation: " + r.Text
	if len(r.Predictions) > 0 {
		s += "\n"
		for _, pred := range r.Predictions {
			s += "\n" + pred.String()
		}
	} else if r.Meaning != nil {
		for _, v := range r.Meaning.Vars {
			s += " " + v.String()
		}
	}
	return s
}

type predByWeight []*Prediction

func (a predByWeight) Len() int           { return len(a) }
func (a predByWeight) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a predByWeight) Less(i, j int) bool { return a[i].Weight > a[j].Weight }

func (r *ParseResult) Merge(other *ParseResult, weight float64) {
	if other.Predictions != nil {
		for _, pred := range other.Predictions {
			pred.Weight *= weight

			r.Predictions = append(r.Predictions, pred)
		}
		sort.Sort(predByWeight(r.Predictions))
	}
}

type Context struct {
	Text          string
	ExpectedReply string
	Sender        string
	Recipient     string
}

type Prediction struct {
	Type string `json:"type"`

	Action  string        `json:"action"`
	Vars    []*Var        `json:"vars"`
	Message proto.Message `json:"msg"`
	Weight  float64       `json:"weight"`
}

func (p Prediction) String() string {
	s := "Intent: " + p.Message.Action
	v := make(map[string]string)
	p.Message.DecodePayload(&v)
	for name, val := range v {
		s += " " + name + "=" + val
	}
	s += "\n       "
	for _, v := range p.Vars {
		s += " " + v.String()
	}
	s += fmt.Sprintf(" [type: %s] [weight: %g]", p.Type, p.Weight)
	return s
}

func (p *Parser) Parse(ctx *Context) (*ParseResult, error) {
	if ctx == nil {
		ctx = &Context{}
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
		r.Predictions = []*Prediction{{Type: typ, Weight: w}}
		return r, nil
	}

	p.pos.Predict(r.Tokens)
	p.ResolvePronouns(r.Tokens, ctx)

	typ := AnalyzeSentenceFunction(r.Tokens)
	var msg proto.Message
	var err error
	switch typ {
	case "declarative":
		if r.Meaning, err = p.meaning.ParseDeclarative(r.Tokens); err != nil {
			return r, err
		}
		action := "event"
		msg, err = p.InventMessageForMeaning(action, r.Meaning)
		r.Predictions = []*Prediction{{
			Type: typ,

			Action:  msg.Action,
			Message: msg,
			Vars:    r.Meaning.Vars,
			Weight:  1, // TODO
		}}

	case "exclamatory":
		r.Predictions = []*Prediction{{
			Type:   typ,
			Weight: 1,
		}}
	case "interrogative":
		if r.Meaning, err = p.meaning.ParseInterrogative(r.Tokens); err != nil {
			return r, err
		}
		msg, err = p.InventMessageForMeaning("concepts/query", r.Meaning)
		r.Predictions = []*Prediction{{
			Type: typ,

			Action:  msg.Action,
			Message: msg,
			Vars:    r.Meaning.Vars,
			Weight:  1, // TODO
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

			fp := &Prediction{
				Type: typ,

				Action: string(pred.Class),
				Weight: float64(pred.Weight) / 10, // TODO: Weights
				Vars:   r.Meaning.Vars,
			}
			r.Predictions = append(r.Predictions, fp)

			if mschema := p.schema.Get(fp.Action); mschema != nil {
				vars := p.varPredictor.PredictTokens(r.Tokens, mschema.Action)
				fp.Vars = append(fp.Vars, vars...)
				fp.Message = mschema.Apply(fp.Vars)
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

func (p *Parser) ReinforceSentence(text string, action string) error {
	r, err := p.Parse(&Context{Text: text})
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
