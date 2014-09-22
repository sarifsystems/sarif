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
)

var Module = core.Module{
	Name:        "events",
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
	DB    Database
	ctx   *core.Context
	proto *proto.Client
}

func NewService(ctx *core.Context) (*Service, error) {
	s := &Service{
		DB:    &sqlDatabase{ctx.Database.Driver(), ctx.Database.DB},
		ctx:   ctx,
		proto: proto.NewClient("events", ctx.Proto),
	}
	return s, nil
}

func (s *Service) Enable() error {
	if err := s.DB.Setup(); err != nil {
		return err
	}
	if err := s.proto.Subscribe("event/new", "", s.handleEventNew); err != nil {
		return err
	}
	if err := s.proto.Subscribe("event/last", "", s.handleEventLast); err != nil {
		return err
	}
	if err := s.proto.Subscribe("location/fence", "", s.handleLocationFence); err != nil {
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
		s.ctx.Log.Warnln("[events] received bad payload:", err)
		s.publish(msg.Reply(proto.BadRequest(err)))
		return
	}
	fixEvent(&e)

	s.ctx.Log.Infoln("[events] new event:", e)

	if err := s.DB.StoreEvent(e); err != nil {
		s.ctx.Log.Errorln("[vents] could not store event:", err)
		s.publish(msg.Reply(proto.InternalError(err)))
		return
	}

	reply := proto.Message{Action: "event/created"}
	if err := reply.EncodePayload(e); err != nil {
		s.ctx.Log.Errorln("[events] could not encode reply:", err)
		s.publish(msg.Reply(proto.InternalError(err)))
		return
	}
	reply.Text = "New event: " + e.String()

	s.publish(msg.Reply(reply))
}

func (s *Service) handleEventLast(msg proto.Message) {
	var filter Event
	if err := msg.DecodePayload(&filter); err != nil {
		s.ctx.Log.Warnln("[events] received bad payload:", err)
		s.publish(msg.Reply(proto.BadRequest(err)))
		return
	}
	fixEvent(&filter)

	s.ctx.Log.Infoln("[events] get last by filter:", filter)
	last, err := s.DB.GetLastEvent(filter)
	if err != nil {
		if err == sql.ErrNoRows {
			s.publish(msg.Reply(MessageEventNotFound))
			return
		}
		s.ctx.Log.Errorln("[events] could not get events:", err)
		s.publish(msg.Reply(proto.InternalError(err)))
		return
	}
	s.ctx.Log.Infoln("[events] found", last)

	reply := proto.Message{Action: "event/found"}
	if err := reply.EncodePayload(last); err != nil {
		s.ctx.Log.Errorln("[events] could not encode reply:", err)
		s.publish(msg.Reply(proto.InternalError(err)))
		return
	}
	reply.Text = last.String()

	s.publish(msg.Reply(reply))
}

func (s *Service) publish(msg proto.Message) {
	if err := s.proto.Publish(msg); err != nil {
		s.ctx.Log.Errorln(err)
	}
}
