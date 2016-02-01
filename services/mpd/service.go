// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service mpd can control the Music Player Daemon.
package mpd

import (
	"github.com/fhs/gompd/mpd"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "mpd",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	Log proto.Logger
	Mpd *mpd.Client
	*proto.Client
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Log:    deps.Log,
		Client: deps.Client,
	}
	return s
}

func (s *Service) Enable() (err error) {
	s.Mpd, err = mpd.Dial("tcp", "localhost:6600")
	if err != nil {
		return err
	}

	s.Subscribe("mpd/play", "", s.handleSimple(s.play))
	s.Subscribe("mpd/pause", "", s.handleSimple(s.pause))
	s.Subscribe("mpd/next", "", s.handleSimple(s.Mpd.Next))
	s.Subscribe("mpd/prev", "", s.handleSimple(s.Mpd.Previous))

	return nil
}

func (s *Service) handleSimple(f func() error) func(proto.Message) {
	return func(msg proto.Message) {
		if err := f(); err != nil {
			s.ReplyInternalError(msg, err)
			return
		}
		s.Reply(msg, proto.CreateMessage("ack/"+msg.Action, nil))
	}
}

func (s *Service) play() error {
	return s.Mpd.Pause(false)
}

func (s *Service) pause() error {
	return s.Mpd.Pause(true)
}
