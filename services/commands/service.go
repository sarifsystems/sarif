// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service commands provides a variety of simple functions.
package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sarifsystems/sarif/pkg/schema"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
)

var Module = &services.Module{
	Name:        "commands",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	Log    sarif.Logger
	Client *sarif.Client
}

type Service struct {
	Log sarif.Logger
	*sarif.Client
}

func NewService(deps *Dependencies) *Service {
	return &Service{
		Log:    deps.Log,
		Client: deps.Client,
	}
}

func (s *Service) Enable() error {
	s.Subscribe("", "commands", s.handleUnknown)
	s.Subscribe("question/answer", "commands", s.handleQuestionAnswer)
	s.Subscribe("cmd/qr", "", s.handleQR)
	s.Subscribe("cmd/increment", "", s.handleCounter)
	s.Subscribe("cmd/decrement", "", s.handleCounter)
	s.Subscribe("cmd/count", "", s.handleCounter)
	s.Subscribe("cmd/ask", "", s.handleAsk)
	s.Subscribe("cmd/catfacts", "", s.handleCatFacts)

	s.Subscribe("cmd/date", "", s.handleDate)
	s.Subscribe("cmd/unix", "", s.handleUnix)
	return nil
}

func (s *Service) ReplyText(orig sarif.Message, text string) error {
	return s.Reply(orig, sarif.Message{
		Action: "ack/" + orig.Action,
		Text:   text,
	})
}

func (s *Service) handleQR(msg sarif.Message) {
	if msg.Text == "" {
		s.ReplyBadRequest(msg, errors.New("No data for QR code specified!"))
		return
	}

	qr := "https://chart.googleapis.com/chart?chs=178x178&cht=qr&chl=" + url.QueryEscape(msg.Text)
	reply := sarif.CreateMessage("ack", map[string]string{
		"url":  qr,
		"type": "image/png",
	})
	reply.Text = qr
	s.Reply(msg, reply)
	return
}

type counterMessage struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func (c counterMessage) Text() string {
	return fmt.Sprintf("Counter '%s' has value %d.", c.Name, c.Value)
}

func (s *Service) handleCounter(msg sarif.Message) {
	if msg.Text == "" {
		s.ReplyBadRequest(msg, errors.New("Please specify a counter name!"))
		return
	}
	name := msg.Text
	cnt, err := s.counterGet(name)
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	newCnt := cnt
	if msg.IsAction("cmd/decrement") {
		newCnt--
	} else if msg.IsAction("cmd/increment") {
		newCnt++
	}
	if newCnt != cnt {
		if err := s.counterSet(name, newCnt); err != nil {
			s.ReplyInternalError(msg, err)
			return
		}
	}
	s.Reply(msg, sarif.CreateMessage("counter/changed/"+name, &counterMessage{name, newCnt}))
}

func (s *Service) counterGet(name string) (int, error) {
	curr, ok := <-s.Request(sarif.CreateMessage("store/get/counter/"+name, nil))
	if !ok {
		return 0, errors.New("Timeout while getting current value")
	}
	if curr.IsAction("err") {
		if curr.IsAction("err/notfound") {
			return 0, nil
		}
		return 0, errors.New(curr.Text)
	}
	var cnt int
	if err := curr.DecodePayload(&cnt); err != nil {
		return 0, err
	}
	return cnt, nil
}

func (s *Service) counterSet(name string, cnt int) error {
	ack, ok := <-s.Request(sarif.CreateMessage("store/put/counter/"+name, cnt))
	if !ok {
		return errors.New("Timeout while setting new value")
	}
	if ack.IsAction("err") {
		return errors.New(ack.Text)
	}
	return nil
}

func (s *Service) handleUnknown(msg sarif.Message) {
	s.Log.Warnln("received unknown message:", msg)
}

type questionMessage struct {
	Question string      `json:"question"`
	Action   interface{} `json:"action"`
}

func (msg questionMessage) Text() string {
	return msg.Question
}

func (s *Service) handleAsk(msg sarif.Message) {
	q := msg.Text
	if q == "" {
		q = "What is the answer to life, the universe and everything?"
	}

	pl := schema.Fill(&questionMessage{
		Question: q,
		Action: schema.Fill(&schema.TextEntryAction{
			Reply: "question/answer/ultimate",
			Name:  "Answer this question.",
		}),
	})
	s.Reply(msg, sarif.CreateMessage("question", pl))
}

func (s *Service) handleQuestionAnswer(msg sarif.Message) {
	if msg.IsAction("question/answer/ultimate") {
		reply := "Wrong. The answer is 42."
		if msg.Text == "42" {
			reply = "Precisely."
		}
		s.Reply(msg, sarif.Message{
			Action: "question/accepted",
			Text:   reply,
		})
		return
	}

	s.Reply(msg, sarif.Message{
		Action: "err/question/unknown",
		Text:   "I can't remember asking you a question.",
	})
}

type CatFactsResponse struct {
	Facts   []string `json:"facts"`
	Success bool     `json:"success,string"`
}

func (r CatFactsResponse) String() string {
	if len(r.Facts) > 0 {
		return r.Facts[0]
	}
	return ""
}

func (s *Service) handleCatFacts(msg sarif.Message) {
	resp, err := http.Get("http://catfacts-api.appspot.com/api/facts")
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	var r CatFactsResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	s.Reply(msg, sarif.CreateMessage("catfacts", &r))
}
