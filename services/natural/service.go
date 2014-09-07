// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/proto/natural"
)

var Module = core.Module{
	Name:        "natural",
	Version:     "1.0",
	NewInstance: newInstance,
}

func newInstance(ctx *core.Context) (core.ModuleInstance, error) {
	s, err := NewService(ctx)
	return s, err
}

func init() {
	core.RegisterModule(Module)
}

type Service struct {
	ctx   *core.Context
	proto *proto.Client
}

func NewService(ctx *core.Context) (*Service, error) {
	s := &Service{
		ctx,
		nil,
	}
	return s, nil
}

func (s *Service) Enable() error {
	s.proto = proto.NewClient("natural", s.ctx.Proto)
	if err := s.proto.Subscribe("natural/handle", "", s.handleNatural); err != nil {
		return err
	}
	if err := s.proto.Subscribe("natural/parse", "", s.handleNaturalParse); err != nil {
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
	text := msg.PayloadGetString("text")
	if text == "" {
		return proto.Message{}, false
	}

	if parsed, ok := natural.ParseRegular(text); ok {
		return parsed, true
	}
	return natural.ParseSimple(text)
}

func (s *Service) handleNatural(msg proto.Message) {
	parsed, ok := s.parseNatural(msg)
	if !ok {
		text := msg.PayloadGetString("text")
		reply := proto.CreateMessage("err/natural", MsgErrNatural{text})
		s.proto.Publish(msg.Reply(reply))
		return
	}

	parsed.Source = msg.Source
	s.proto.Publish(parsed)
}

func (s *Service) handleNaturalParse(msg proto.Message) {
	text := msg.PayloadGetString("text")
	parsed, ok := s.parseNatural(msg)
	if !ok {
		reply := proto.CreateMessage("err/natural", MsgErrNatural{text})
		s.proto.Publish(msg.Reply(reply))
		return
	}

	reply := proto.CreateMessage("natural/parsed", MsgNaturalParsed{parsed, text})
	s.proto.Publish(msg.Reply(reply))
}
