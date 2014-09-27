// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package know

import (
	"github.com/xconstruct/know"
	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
)

var Module = core.Module{
	Name:        "know",
	Version:     "1.0",
	NewInstance: newInstance,
}

func init() {
	core.RegisterModule(Module)
}

func newInstance(ctx *core.Context) (core.ModuleInstance, error) {
	return NewService(ctx)
}

type Service struct {
	ctx   *core.Context
	proto *proto.Client
}

func NewService(ctx *core.Context) (*Service, error) {
	s := &Service{
		ctx:   ctx,
		proto: proto.NewClient("know", ctx.Proto),
	}
	return s, nil
}

func (s *Service) Enable() error {
	if err := s.proto.Subscribe("knowledge/query", "", s.handleQuery); err != nil {
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
	if m.Answer == "" {
		return "No answer found for '" + m.Query + "'."
	}

	return m.Query + " is " + m.Answer + "."
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
			s.proto.Publish(msg.Reply(proto.CreateMessage("knowledge/noanswer", pl)))
			return
		}

		// Error received, forward.
		s.proto.Publish(msg.Reply(proto.InternalError(err)))
		return
	}

	// Send answer.
	pl := MessageAnswer{
		ans.Question,
		ans.Answer,
		ans.Provider,
	}
	s.proto.Publish(msg.Reply(proto.CreateMessage("knowledge/answer", pl)))
	return
}
