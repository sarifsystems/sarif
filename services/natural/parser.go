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
	Parser *natural.Model
	Pos    *mlearning.Model
}

type Parser struct {
	tokenizer *natural.Tokenizer
	parser    *natural.LearningParser
	pos       *natural.PosTagger
	meaning   *natural.MeaningParser
}

func NewParser() *Parser {
	return &Parser{
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
			natural.TrainDefaults(p.parser)
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
}

func (p *Parser) Parse(text string, ctx *Context) (*ParseResult, error) {
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

	r.Tokens = p.tokenizer.Tokenize(text)

	if ctx.ExpectedReply == "affirmative" {
		r.Type, r.Weight = natural.AnalyzeAffirmativeSentiment(r.Tokens)
		return r, nil
	}

	p.pos.Predict(r.Tokens)

	// TODO
	r.Type = natural.AnalyzeSentenceFunction(r.Tokens)
	var err error
	switch r.Type {
	case "declarative":
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

	r.Message, r.Weight = p.parser.Parse(text)
	return r, err
}

func inStringSlice(s string, ss []string) bool {
	for _, a := range ss {
		if s == a {
			return true
		}
	}
	return false
}
