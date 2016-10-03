// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/sarifsystems/sarif/pkg/natural"
	"github.com/sarifsystems/sarif/pkg/schema"
	"github.com/sarifsystems/sarif/sarif"
)

type Conversation struct {
	service *Service

	Device            string
	LastTime          time.Time
	LastMessage       sarif.Message
	LastMessageAction Actionable

	LastUserTime    time.Time
	LastUserText    string
	LastUserMessage sarif.Message
}

type MsgErrNatural struct {
	Original string               `json:"original"`
	Type     string               `json:"-"`
	Action   interface{}          `json:"action"`
	Result   *natural.ParseResult `json:"result"`
}

type Actionable struct {
	Action  *schema.Action   `json:"action"`
	Actions []*schema.Action `json:"actions"`
}

func (a Actionable) IsAction() bool {
	if a.Action == nil || a.Action.Thing == nil {
		return false
	}
	return strings.HasSuffix(a.Action.SchemaType, "Action")
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

func (cv *Conversation) PublishForClient(msg sarif.Message) {
	msg.Source = cv.service.DeviceId + "/" + cv.Device
	cv.service.Publish(msg)
}

func (cv *Conversation) SendToClient(msg sarif.Message) {
	// Save conversation.
	cv.LastTime = time.Now()
	cv.LastMessage = msg

	// Analyze message for possible user actions
	cv.LastMessageAction = Actionable{}
	msg.DecodePayload(&cv.LastMessageAction)

	msg = cv.service.AnnotateReply(msg)

	// Forward response to client.
	msg.Id = sarif.GenerateId()
	msg.Destination = cv.Device
	cv.service.Publish(msg)
}

func (cv *Conversation) HandleClientMessage(msg sarif.Message) {
	if msg.Text == ".full" || msg.Text == "/full" {
		text, err := json.MarshalIndent(cv.LastMessage, "", "    ")
		if err != nil {
			panic(err)
		}
		cv.service.Reply(msg, sarif.Message{
			Action: "natural/full",
			Text:   string(text),
		})
		return
	}

	// Check if client answers a conversation.
	if time.Now().Sub(cv.LastTime) < 5*time.Minute {
		if cv.LastMessageAction.IsAction() {
			parsed, ok := cv.answer(cv.LastMessageAction.Action, msg.Text)
			cv.LastTime = time.Time{}
			parsed.Destination = cv.LastMessage.Source
			if ok {
				cv.PublishForClient(parsed)
			}
			return
		}
	}

	// Otherwise parse message as normal request.
	ctx := &natural.Context{
		Text:      msg.Text,
		Sender:    "user",
		Recipient: "sarif",
	}
	res, err := cv.service.Parse(ctx)
	if err != nil || len(res.Intents) == 0 {
		cv.handleUnknownUserMessage(msg)
		return
	}
	pred := res.Intents[0]
	if pred.Type == "exclamatory" {
		cv.SendToClient(msg.Reply(sarif.Message{
			Action: "natural/phrase",
			Text:   cv.service.phrases.Answer(msg.Text),
		}))
		return
	}

	if pred.Message.Text == "" && pred.Type != "simple" {
		pred.Message.Text = msg.Text
	}
	cv.LastUserTime = time.Now()
	cv.LastUserText = msg.Text
	cv.LastUserMessage = pred.Message
	pred.Message.CorrId = msg.Id
	cv.PublishForClient(pred.Message)
}

func (cv *Conversation) answer(a *schema.Action, text string) (sarif.Message, bool) {
	reply := sarif.Message{
		Action: a.Reply,
		Text:   text,
	}
	if text == ".cancel" || text == "cancel" || strings.HasPrefix(text, "cancel ") {
		return reply, false
	}

	t := a.SchemaType
	if t == "ConfirmAction" || t == "DeleteAction" || t == "CancelAction" {
		ctx := &natural.Context{Text: text, ExpectedReply: "affirmative"}
		r, err := cv.service.Parse(ctx)
		if err != nil || len(r.Intents) == 0 {
			return reply, false
		}
		if r.Intents[0].Type == "neg" {
			reply.Action = a.ReplyNegative
			return reply, reply.Action != ""
		}
	}

	if a.Payload != nil {
		reply.EncodePayload(a.Payload)
	}
	return reply, true
}

func (cv *Conversation) handleUnknownUserMessage(msg sarif.Message) {
	pl := &MsgErrNatural{
		Original: msg.Text,
	}

	cv.SendToClient(msg.Reply(sarif.CreateMessage("err/natural", pl)))
}
