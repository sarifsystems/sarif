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
	s.Subscribe("natural/learn", "", s.handleNaturalLearn)
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
	res, err := s.parser.Parse(msg.Text, &Context{
		Sender:    "user",
		Recipient: "stark",
	})
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

type msgLearn struct {
	Sentence string `json:"sentence"`
	Action   string `json:"action"`
}

func (p msgLearn) Text() string {
	return "I learned to associate '" + p.Sentence + "' with " + p.Action + "."
}

func (s *Service) handleNaturalLearn(msg proto.Message) {
	var p msgLearn
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	p.Sentence = strings.TrimSpace(p.Sentence)
	if p.Sentence == "" {
		return
	}
	p.Action = strings.TrimLeft(p.Action, ".")

	if err := s.parser.regular.Learn(p.Sentence, p.Action); err != nil {
		s.ReplyBadRequest(msg, err)
	}
	s.Log.Infof("[natural] associating '%s' with %s", p.Sentence, p.Action)
	s.Reply(msg, proto.CreateMessage("natural/learned/meaning", p))
}

func (s *Service) handleNaturalReinforce(msg proto.Message) {
	var p msgLearn
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	if p.Sentence == "" || p.Action == "" {
		s.ReplyBadRequest(msg, errors.New("No sentence or action given."))
		return
	}

	s.Log.Infof("[natural] reinforcing: '%s' with %s", p.Sentence, p.Action)
	s.parser.ReinforceSentence(p.Sentence, p.Action)

	parsed, _ := s.parser.Parse(p.Sentence, &Context{})
	s.Reply(msg, proto.CreateMessage("natural/learned/meaning", &msgLearn{p.Sentence, parsed.Message.Action}))
}
