// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package nlparser

import (
	"errors"
	"time"

	"github.com/sarifsystems/sarif/pkg/natural"
	"github.com/sarifsystems/sarif/pkg/natural/nlp"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
)

var Module = &services.Module{
	Name:        "nlparser",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	ModelPath string
}

type Dependencies struct {
	Config services.Config
	Client sarif.Client
}

type Service struct {
	Config services.Config
	Cfg    Config
	sarif.Client

	parser *nlp.Parser
}

func NewService(deps *Dependencies) *Service {
	return &Service{
		deps.Config,
		Config{},
		deps.Client,

		nlp.NewParser(),
	}
}

func (s *Service) Enable() error {
	s.Subscribe("natural/parse", "self", s.handleNaturalParse)
	s.Subscribe("natural/reinforce", "self", s.handleNaturalReinforce)

	s.Cfg.ModelPath = s.Config.Dir() + "/" + "natural.json.gz"
	s.Config.Get(&s.Cfg)

	// TODO: Use a more idiomatic config way to disable filesystem access
	if s.Config.Dir() != "" {
		go func() {
			if err := s.loadModel(); err != nil {
				s.Log("err", err.Error())
				return
			}
			s.saveModelLoop()
		}()
	}
	return nil
}
func (s *Service) saveModelLoop() {
	time.Sleep(1 * time.Minute)
	for {
		if err := s.saveModel(); err != nil {
			s.Log("err/internal", "error saving model: "+err.Error())
		}
		time.Sleep(time.Hour)
	}
}

func (s *Service) loadModel() error {
	s.Log("debug", "loading model from "+s.Cfg.ModelPath)
	return s.parser.LoadModel(s.Cfg.ModelPath)
}

func (s *Service) saveModel() error {
	s.Log("debug", "saving model to "+s.Cfg.ModelPath)
	return s.parser.SaveModel(s.Cfg.ModelPath)
}

func (s *Service) handleNaturalParse(msg sarif.Message) {
	ctx := &natural.Context{}
	msg.DecodePayload(ctx)
	if ctx.Text == "" {
		ctx.Text = msg.Text
	}

	res, err := s.parser.Parse(ctx)
	if err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	s.Reply(msg, sarif.CreateMessage("natural/parsed", res))
}

type msgLearn struct {
	Sentence string `json:"sentence"`
	Action   string `json:"action"`
}

func (p msgLearn) Text() string {
	return "I learned to associate '" + p.Sentence + "' with " + p.Action + "."
}

func (s *Service) handleNaturalReinforce(msg sarif.Message) {
	var p msgLearn
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	if p.Sentence == "" || p.Action == "" {
		s.ReplyBadRequest(msg, errors.New("No sentence or action given."))
		return
	}

	s.Log("info", "reinforcing '"+p.Sentence+"' with "+p.Action)
	s.parser.ReinforceSentence(p.Sentence, p.Action)

	parsed, _ := s.parser.Parse(&natural.Context{Text: p.Sentence})
	s.Reply(msg, sarif.CreateMessage("natural/learned/meaning", &msgLearn{p.Sentence, parsed.Intents[0].Message.Action}))
}
