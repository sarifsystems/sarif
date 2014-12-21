// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"encoding/json"
	"fmt"
	"strings"
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
	Device      string
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
	if err := s.Subscribe("natural/handle", "", s.handleNatural); err != nil {
		return err
	}
	if err := s.Subscribe("natural/parse", "", s.handleNaturalParse); err != nil {
		return err
	}
	if err := s.Subscribe("", "user", s.handleUserMessage); err != nil {
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
		cv = &Conversation{
			Device: device,
		}
		s.conversations[device] = cv
		s.Subscribe("", s.DeviceId+"/"+device, s.handleNetworkMessage)
	}
	return cv
}

type Actionable struct {
	Action  *schema.Action   `json:"action"`
	Actions []*schema.Action `json:"actions"`
}

func (a Actionable) IsAction() bool {
	if a.Action == nil {
		return false
	}
	if a.Action.SchemaType == "TextEntryAction" {
		return true
	}

	return false
}

func answer(a *schema.Action, text string) (proto.Message, bool) {
	reply := proto.Message{
		Action: a.Reply,
		Text:   text,
	}
	if text == ".cancel" || text == "cancel" || strings.HasPrefix(text, "cancel ") {
		return reply, false
	}
	if a.SchemaType == "TextEntryAction" {
		return reply, true
	}
	return reply, false
}

func (s *Service) publishForClient(cv *Conversation, msg proto.Message) {
	msg.Source = s.DeviceId + "/" + cv.Device
	s.Publish(msg)
}

func (s *Service) forwardToClient(cv *Conversation, msg proto.Message) {
	// Save conversation.
	cv.LastTime = time.Now()
	cv.LastMessage = msg

	// Forward response to client.
	msg.Id = proto.GenerateId()
	msg.Destination = cv.Device
	s.Publish(msg)
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

	// Check if client answers a conversation.
	if time.Now().Sub(cv.LastTime) < 5*time.Minute {
		var act Actionable
		if err := cv.LastMessage.DecodePayload(&act); err == nil {
			if act.IsAction() {
				parsed, ok := answer(act.Action, msg.Text)
				parsed.Destination = cv.LastMessage.Source
				cv.LastTime = time.Time{}
				if ok {
					s.publishForClient(cv, parsed)
				}
				return
			}
		}
	}

	// Otherwise parse message as normal request.
	parsed, ok := s.parseNatural(msg)
	if !ok {
		reply := proto.CreateMessage("err/natural", MsgErrNatural{msg.Text})
		s.Publish(msg.Reply(reply))
		return
	}
	s.publishForClient(cv, parsed)
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

func (s *Service) handleNetworkMessage(msg proto.Message) {
	client := strings.TrimPrefix(msg.Destination, s.DeviceId+"/")
	fmt.Println(client, "moo")
	cv := s.getConversation(client)
	s.forwardToClient(cv, msg)
}

func (s *Service) handleUserMessage(msg proto.Message) {
	for _, cv := range s.conversations {
		s.forwardToClient(cv, msg)
	}
}
