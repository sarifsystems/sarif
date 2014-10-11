// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package events

import (
	"database/sql"
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "events",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	DB     *core.DB
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	DB  Database
	Log proto.Logger
	*proto.Client
}

func NewService(deps *Dependencies) *Service {
	return &Service{
		DB:     &sqlDatabase{deps.DB.Driver(), deps.DB.DB},
		Log:    deps.Log,
		Client: deps.Client,
	}
}

func (s *Service) Enable() error {
	if err := s.DB.Setup(); err != nil {
		return err
	}
	if err := s.Subscribe("event/new", "", s.handleEventNew); err != nil {
		return err
	}
	if err := s.Subscribe("event/last", "", s.handleEventLast); err != nil {
		return err
	}
	if err := s.Subscribe("location/fence", "", s.handleLocationFence); err != nil {
		return err
	}
	return nil
}

func (s *Service) Disable() error { return nil }

var MessageEventNotFound = proto.Message{
	Action: "event/notfound",
	Text:   "No event found.",
}

func fixEvent(e *Event) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	if e.Subject == "i" || e.Subject == "I" {
		e.Subject = "user"
	}
}

func (s *Service) handleEventNew(msg proto.Message) {
	var e Event
	if err := msg.DecodePayload(&e); err != nil {
		s.Log.Warnln("[events] received bad payload:", err)
		s.ReplyBadRequest(msg, err)
		return
	}
	fixEvent(&e)
	if e.Text == "" {
		e.Text = msg.Text
	}

	s.Log.Infoln("[events] new event:", e)

	if err := s.DB.StoreEvent(e); err != nil {
		s.Log.Errorln("[vents] could not store event:", err)
		s.ReplyInternalError(msg, err)
		return
	}

	reply := proto.Message{Action: "event/created"}
	if err := reply.EncodePayload(e); err != nil {
		s.Log.Errorln("[events] could not encode reply:", err)
		s.ReplyInternalError(msg, err)
		return
	}
	reply.Text = "New event: " + e.String()

	s.Reply(msg, reply)
}

func (s *Service) handleEventLast(msg proto.Message) {
	var filter Event
	if err := msg.DecodePayload(&filter); err != nil {
		s.Log.Warnln("[events] received bad payload:", err)
		s.ReplyBadRequest(msg, err)
		return
	}
	fixEvent(&filter)

	s.Log.Infoln("[events] get last by filter:", filter)
	last, err := s.DB.GetLastEvent(filter)
	if err != nil {
		if err == sql.ErrNoRows {
			s.Reply(msg, MessageEventNotFound)
			return
		}
		s.Log.Errorln("[events] could not get events:", err)
		s.ReplyInternalError(msg, err)
		return
	}
	s.Log.Infoln("[events] found", last)

	reply := proto.Message{Action: "event/found"}
	if err := reply.EncodePayload(last); err != nil {
		s.Log.Errorln("[events] could not encode reply:", err)
		s.ReplyInternalError(msg, err)
		return
	}
	reply.Text = last.String()

	s.Reply(msg, reply)
}
