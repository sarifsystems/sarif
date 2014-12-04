// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package commands

import (
	"errors"
	"net/url"

	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "commands",
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
		Log:    deps.Log,
		Client: deps.Client,
	}
}

func (s *Service) Enable() error {
	if err := s.Subscribe("cmd/qr", "", s.handleQR); err != nil {
		return err
	}
	return nil
}

func (s *Service) handleQR(msg proto.Message) {
	if msg.Text == "" {
		s.ReplyBadRequest(msg, errors.New("No data for QR code specified!"))
		return
	}

	qr := "https://chart.googleapis.com/chart?chs=178x178&cht=qr&chl=" + url.QueryEscape(msg.Text)
	reply := proto.CreateMessage("ack", map[string]string{
		"url": qr,
	})
	reply.Text = qr
	s.Reply(msg, reply)
	return
}
