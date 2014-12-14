// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/xconstruct/stark/pkg/natural"
	"github.com/xconstruct/stark/pkg/schema"
	"github.com/xconstruct/stark/proto"
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

type Conversation struct {
	LastTime    time.Time
	LastMessage proto.Message
}

type Service struct {
	Log proto.Logger
	*proto.Client

	conversations map[string]*Conversation
}

func NewService(deps *Dependencies) *Service {
	return &Service{
		deps.Log,
		deps.Client,

		make(map[string]*Conversation),
	}
}

func (s *Service) Enable() error {
	s.Client.RequestTimeout = 10 * time.Second
	if err := s.Subscribe("natural/handle", "", s.handleNatural); err != nil {
		return err
	}
	if err := s.Subscribe("natural/parse", "", s.handleNaturalParse); err != nil {
		return err
	}
	if err := s.Subscribe("", "self", s.handleUnknownResponse); err != nil {
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

func (s *Service) getConversation(device string) *Conversation {
	cv, ok := s.conversations[device]
	if !ok {
		cv = &Conversation{}
		s.conversations[device] = cv
	}
	return cv
}

type Actionable struct {
	Action  *schema.Action   `json:"action"`
	Actions []*schema.Action `json:"actions"`
}

func (s *Service) handleNatural(msg proto.Message) {
	cv := s.getConversation(msg.Source)

	if msg.Text == ".full" {
		text, err := json.MarshalIndent(cv.LastMessage, "", "    ")
		if err != nil {
			panic(err)
		}
		s.Reply(msg, proto.Message{
			Action: "natural/full",
			Text:   string(text),
		})
		return
	}
	if msg.Text == ".cancel" {
		cv.LastTime = time.Time{}
		return
	}

	var ok bool
	var parsed proto.Message

	// Check if client answers a conversation
	if time.Now().Sub(cv.LastTime) < 5*time.Minute {
		var act Actionable
		if err := cv.LastMessage.DecodePayload(&act); err == nil {
			if act.Action != nil && act.Action.SchemaType == "TextEntryAction" {
				parsed = proto.Message{
					Action:      act.Action.Reply,
					Text:        msg.Text,
					Destination: cv.LastMessage.Source,
				}
				ok = true
				cv.LastTime = time.Time{}
			}
		}
	}

	if !ok {
		parsed, ok = s.parseNatural(msg)
	}

	if !ok {
		reply := proto.CreateMessage("err/natural", MsgErrNatural{msg.Text})
		s.Publish(msg.Reply(reply))
		return
	}

	if msg.IsAction("natural/handle/direct") {
		parsed.Source = msg.Source
		return
	}

	// Execute parsed message for the client and retrieve responses.
	for reply := range s.Request(parsed) {
		// Save conversation.
		cv.LastTime = time.Now()
		cv.LastMessage = reply

		// Forward response to client.
		reply.Id = proto.GenerateId()
		reply.Destination = msg.Source
		reply.CorrId = msg.Id
		s.Publish(reply)
	}
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

func (s *Service) handleUnknownResponse(msg proto.Message) {
	s.ReplyBadRequest(msg, errors.New("Received unknown reply."))
}
