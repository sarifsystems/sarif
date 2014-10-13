// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/proto/natural"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "natural",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	Log proto.Logger
	*proto.Client
}

func NewService(deps *Dependencies) *Service {
	return &Service{
		deps.Log,
		deps.Client,
	}
}

func (s *Service) Enable() error {
	if err := s.Subscribe("natural/handle", "", s.handleNatural); err != nil {
		return err
	}
	if err := s.Subscribe("natural/parse", "", s.handleNaturalParse); err != nil {
		return err
	}
	return nil
}

func (s *Service) Disable() error {
	return nil
}

type MsgErrNatural struct {
	Original string `json:"original"`
}

func (pl MsgErrNatural) String() string {
	return "I didn't understand your message."
}

type MsgNaturalParsed struct {
	Parsed   proto.Message
	Original string `json:"original"`
}

func (pl MsgNaturalParsed) String() string {
	return "Natural message correctly parsed."
}

func (s *Service) parseNatural(msg proto.Message) (proto.Message, bool) {
	if msg.Text == "" {
		return proto.Message{}, false
	}

	parsed, ok := natural.ParseRegular(msg.Text)
	if !ok {
		parsed, ok = natural.ParseSimple(msg.Text)
	}
	if !ok {
		return parsed, false
	}

	return parsed, true
}

func (s *Service) handleNatural(msg proto.Message) {
	parsed, ok := s.parseNatural(msg)
	if !ok {
		reply := proto.CreateMessage("err/natural", MsgErrNatural{msg.Text})
		s.Publish(msg.Reply(reply))
		return
	}

	parsed.Source = msg.Source
	s.Publish(parsed)
}

func (s *Service) handleNaturalParse(msg proto.Message) {
	parsed, ok := s.parseNatural(msg)
	if !ok {
		reply := proto.CreateMessage("err/natural", MsgErrNatural{msg.Text})
		s.Publish(msg.Reply(reply))
		return
	}

	reply := proto.CreateMessage("natural/parsed", MsgNaturalParsed{parsed, msg.Text})
	s.Publish(msg.Reply(reply))
}
