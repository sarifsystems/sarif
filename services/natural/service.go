// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/xconstruct/stark/pkg/natural"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "natural",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	Config services.Config
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	Config services.Config
	Log    proto.Logger
	*proto.Client

	parser        *Parser
	conversations map[string]*Conversation
}

func NewService(deps *Dependencies) *Service {
	return &Service{
		deps.Config,
		deps.Log,
		deps.Client,

		NewParser(),
		make(map[string]*Conversation),
	}
}

func (s *Service) Enable() error {
	s.Subscribe("natural/handle", "", s.handleNatural)
	s.Subscribe("natural/parse", "", s.handleNaturalParse)
	s.Subscribe("natural/learn/sentence", "", s.handleNaturalLearnSentence)
	s.Subscribe("natural/learn/meaning", "", s.handleNaturalLearnMeaning)
	s.Subscribe("natural/reinforce", "", s.handleNaturalReinforce)
	s.Subscribe("", "user", s.handleUserMessage)

	if err := s.loadModel(); err != nil {
		return err
	}
	go s.saveModelLoop()
	return nil
}

func (s *Service) saveModelLoop() {
	time.Sleep(1 * time.Minute)
	if err := s.saveModel(); err != nil {
		s.Log.Errorln("[natural] error saving model:", err)
		return
	}
	c := time.Tick(time.Hour)
	for n := range c {
		if err := s.saveModel(); err != nil {
			s.Log.Errorln("[natural] error saving model:", err, n)
		}
	}
}

func (s *Service) loadModel() error {
	path := s.Config.Dir() + "/" + "natural.json.gz"
	s.Log.Debugln("[natural] loading model from", path)
	return s.parser.LoadModel(path)
}

func (s *Service) saveModel() error {
	path := s.Config.Dir() + "/" + "natural.json.gz"
	s.Log.Debugln("[natural] saving model to", path)
	return s.parser.SaveModel(path)
}

func (s *Service) Disable() error {
	return nil
}

type MsgNaturalParsed struct {
	*ParseResult
}

func (p MsgNaturalParsed) String() string {
	if p.Weight <= 0 {
		return "Could not parse this text."
	}
	return fmt.Sprintf("Interpreted '%s' as action '%s'.", p.Text, p.Message.Action)
}

func (s *Service) getConversation(device string) *Conversation {
	cv, ok := s.conversations[device]
	if !ok {
		cv = &Conversation{
			service: s,
			Device:  device,
		}
		s.conversations[device] = cv
		s.Subscribe("", s.DeviceId+"/"+device, s.handleNetworkMessage)
	}
	return cv
}

func (s *Service) handleNatural(msg proto.Message) {
	cv := s.getConversation(msg.Source)
	cv.HandleClientMessage(msg)
}

func (s *Service) handleNaturalParse(msg proto.Message) {
	res, err := s.parser.Parse(msg.Text, &Context{})
	if err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	action := "natural/parsed"
	if res.Weight <= 0 {
		action = "err/natural/parsed"
	}

	s.Reply(msg, proto.CreateMessage(action, MsgNaturalParsed{res}))
}

func (s *Service) handleNetworkMessage(msg proto.Message) {
	client := strings.TrimPrefix(msg.Destination, s.DeviceId+"/")
	cv := s.getConversation(client)
	cv.SendToClient(msg)
}

func (s *Service) handleUserMessage(msg proto.Message) {
	for _, cv := range s.conversations {
		cv.SendToClient(msg)
	}
}

type msgLearnedSentence struct {
	Rule   string `json:"rule"`
	Action string `json:"action,omitempty"`
}

func (p msgLearnedSentence) String() string {
	if p.Action != "" {
		return fmt.Sprintf("I learned '%s'. I think it means %s.", p.Rule, p.Action)
	}
	return fmt.Sprintf("I learned '%s'.", p.Rule)
}

func (s *Service) handleNaturalLearnSentence(msg proto.Message) {
	rule := strings.TrimSpace(msg.Text)
	if rule == "" {
		return
	}

	s.Log.Infof("[natural] learning new rule: '%s'", rule)
	s.parser.parser.LearnSentence(rule)

	action := ""
	if res, err := s.parser.Parse(msg.Text, &Context{}); err != nil && res.Weight > 0 {
		action = res.Message.Action
	}
	s.Reply(msg, proto.CreateMessage("natural/learned/sentence", &msgLearnedSentence{rule, action}))
}

type msgLearnMeaning struct {
	Sentence string `json:"sentence"`
}

type msgLearnedMeaning struct {
	Sentence string `json:"sentence"`
	Action   string `json:"action"`
}

func (p msgLearnedMeaning) String() string {
	return "I learned to associate '" + p.Sentence + "' with " + p.Action + "."
}

func (s *Service) handleNaturalLearnMeaning(msg proto.Message) {
	var p msgLearnMeaning
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	sentence := strings.TrimSpace(p.Sentence)
	if sentence == "" {
		return
	}

	var parsed proto.Message
	ok := false
	if parsed, ok = natural.ParseSimple(msg.Text); !ok {
		parsed.Action = strings.TrimLeft(strings.TrimSpace(msg.Text), ".")
	}

	s.Log.Infof("[natural] reinforcing: '%s' with %s", sentence, parsed.Action)
	s.parser.parser.LearnMessage(parsed)
	s.parser.parser.ReinforceSentence(sentence, parsed.Action)
	s.Reply(msg, proto.CreateMessage("natural/learned/meaning", &msgLearnedMeaning{sentence, parsed.Action}))
}

func (s *Service) handleNaturalReinforce(msg proto.Message) {
	var p msgLearnedMeaning
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	if p.Sentence == "" || p.Action == "" {
		s.ReplyBadRequest(msg, errors.New("No sentence or action given."))
		return
	}

	s.Log.Infof("[natural] reinforcing: '%s' with %s", p.Sentence, p.Action)
	s.parser.parser.ReinforceSentence(p.Sentence, p.Action)

	parsed, _ := s.parser.parser.Parse(p.Sentence)
	s.Reply(msg, proto.CreateMessage("natural/learned/meaning", &msgLearnedMeaning{p.Sentence, parsed.Action}))
}
