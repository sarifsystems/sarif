// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/xconstruct/stark/pkg/natural"
	"github.com/xconstruct/stark/pkg/schema"
	"github.com/xconstruct/stark/proto"
)

type Conversation struct {
	service *Service

	Device            string
	LastTime          time.Time
	LastMessage       proto.Message
	LastMessageAction Actionable

	LastUserText    string
	LastUserMessage proto.Message
}

type MsgErrNatural struct {
	Original string      `json:"original"`
	Type     string      `json:"-"`
	Action   interface{} `json:"action"`
}

type Actionable struct {
	Action  *schema.Action   `json:"action"`
	Actions []*schema.Action `json:"actions"`
}

func (a Actionable) IsAction() bool {
	if a.Action == nil || a.Action.Thing == nil {
		return false
	}
	if a.Action.SchemaType == "TextEntryAction" {
		return true
	}

	return false
}

func (pl MsgErrNatural) String() string {
	switch pl.Type {
	case "sentence":
		return "I didn't understand your message. Please give me a rule to learn this sentence."
	case "meaning":
		return "I didn't understand your message. Please give me an example message or action to learn this sentence."
	}
	return "I didn't understand your message."
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
		if a.Payload != nil {
			reply.EncodePayload(a.Payload)
		}
		return reply, true
	}
	return reply, false
}

func (cv *Conversation) PublishForClient(msg proto.Message) {
	msg.Source = cv.service.DeviceId + "/" + cv.Device
	cv.service.Publish(msg)
}

func (cv *Conversation) SendToClient(msg proto.Message) {
	// Save conversation.
	cv.LastTime = time.Now()
	cv.LastMessage = msg

	// Analyze message for possible user actions
	cv.LastMessageAction = Actionable{}
	msg.DecodePayload(&cv.LastMessageAction)

	// Forward response to client.
	msg.Id = proto.GenerateId()
	msg.Destination = cv.Device
	natural.FormatMessage(&msg)
	cv.service.Publish(msg)
}

func (cv *Conversation) HandleClientMessage(msg proto.Message) {
	if msg.Text == ".full" {
		text, err := json.MarshalIndent(cv.LastMessage, "", "    ")
		if err != nil {
			panic(err)
		}
		cv.service.Reply(msg, proto.Message{
			Action: "natural/full",
			Text:   string(text),
		})
		return
	}

	// Check if client answers a conversation.
	if time.Now().Sub(cv.LastTime) < 5*time.Minute {
		if cv.LastMessageAction.IsAction() {
			parsed, ok := answer(cv.LastMessageAction.Action, msg.Text)
			cv.LastTime = time.Time{}
			parsed.Destination = cv.LastMessage.Source
			if ok {
				cv.PublishForClient(parsed)
			}
			return
		}
	}

	// Otherwise parse message as normal request.
	parsed, ok := cv.service.parseNatural(msg)
	if !ok {
		cv.handleUnknownUserMessage(msg)
		return
	}

	cv.LastUserText = msg.Text
	cv.LastUserMessage = parsed
	parsed.CorrId = msg.Id
	cv.PublishForClient(parsed)
}

func (cv *Conversation) handleUnknownUserMessage(msg proto.Message) {
	pl := &MsgErrNatural{
		Original: msg.Text,
	}
	if m := cv.service.parser.ParseSentence(msg.Text); m != nil {
		pl.Type = "meaning"
		pl.Action = schema.Fill(&schema.TextEntryAction{
			Reply:   "natural/learn/meaning",
			Name:    "Give an example message or action to learn this sentence.",
			Payload: &msgLearnMeaning{msg.Text},
		})
	} else {
		pl.Type = "sentence"
		pl.Action = schema.Fill(&schema.TextEntryAction{
			Reply: "natural/learn/sentence",
			Name:  "Give a rule to learn this sentence.",
		})
	}

	cv.SendToClient(msg.Reply(proto.CreateMessage("err/natural", pl)))
}
