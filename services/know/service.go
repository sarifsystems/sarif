// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service know answers questions by asking multiple knowledge providers.
package know

import (
	"strings"

	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
	"github.com/xconstruct/know"
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
	Config services.Config
	Client *sarif.Client
}

type Service struct {
	cfg Config
	Log sarif.Logger
	*sarif.Client
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Client: deps.Client,
	}
	deps.Config.Get(&s.cfg)
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

func (s *Service) handleQuery(msg sarif.Message) {
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
			s.Reply(msg, sarif.CreateMessage("knowledge/noanswer", pl))
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
	s.Reply(msg, sarif.CreateMessage("knowledge/answer", pl))
	return
}
