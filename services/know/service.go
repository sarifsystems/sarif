// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package know

import (
	"strings"

	"github.com/xconstruct/know"
	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "know",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	WolframApiKey string
}

type Dependencies struct {
	Config *core.Config
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	cfg Config
	Log proto.Logger
	*proto.Client
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Log:    deps.Log,
		Client: deps.Client,
	}
	deps.Config.Get("know", &s.cfg)
	if s.cfg.WolframApiKey != "" {
		know.Wolfram.SetApiKey(s.cfg.WolframApiKey)
	}
	return s
}

func (s *Service) Enable() error {
	if err := s.Subscribe("knowledge/query", "", s.handleQuery); err != nil {
		return err
	}
	return nil
}

func (s *Service) Disable() error { return nil }

type MessageAnswer struct {
	Query  string `json:"query"`
	Answer string `json:"answer"`
	Source string
}

func (m MessageAnswer) String() string {
	ans := m.Answer
	if ans == "" {
		return "No answer found for '" + m.Query + "'."
	}
	if strings.Contains(ans, "\n") {
		ans = ":\n" + ans
	}

	return m.Query + " is " + ans + "."
}

func (s *Service) handleQuery(msg proto.Message) {
	query := msg.Text

	// Query and wait for first answer
	answers, errs := know.Ask(query)
	ans, ok := <-answers

	if !ok {
		// No answer found? Check for errors.
		err, ok := <-errs
		if !ok {
			// No errors found? Send negative answer
			pl := MessageAnswer{
				Query: query,
			}
			s.Reply(msg, proto.CreateMessage("knowledge/noanswer", pl))
			return
		}

		// Error received, forward.
		s.ReplyInternalError(msg, err)
		return
	}

	// Send answer.
	pl := MessageAnswer{
		ans.Question,
		ans.Answer,
		ans.Provider,
	}
	s.Reply(msg, proto.CreateMessage("knowledge/answer", pl))
	return
}
