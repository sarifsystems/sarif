// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package events

import (
	"database/sql"

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
	s.proto.RegisterHandler(s.handle)
	return s, nil
}

func (s *Service) Enable() error {
	if err := s.DB.Setup(); err != nil {
		return err
	}
	if err := s.proto.SubscribeGlobal("event"); err != nil {
		return err
	}
	return nil
}
func (s *Service) Disable() error { return nil }

var MessageEventNotFound = proto.Message{
	Action: "event/notfound",
	Payload: map[string]interface{}{
		"text": "No event found.",
	},
}

func (s *Service) handle(msg proto.Message) {
	if !msg.IsAction("event") {
		return
	}

	if msg.IsAction("event/new") {
		var e Event
		if err := msg.DecodePayload(&e); err != nil {
			s.ctx.Log.Warnln("[events] received bad payload:", err)
			s.publish(msg.Reply(proto.BadRequest(err)))
			return
		}

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

		s.publish(msg.Reply(reply))
		return
	}

	if msg.IsAction("event/last") {
		var filter Event
		if err := msg.DecodePayload(&filter); err != nil {
			s.ctx.Log.Warnln("[events] received bad payload:", err)
			s.publish(msg.Reply(proto.BadRequest(err)))
			return
		}

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

		reply := proto.Message{Action: "event/found"}
		if err := reply.EncodePayload(last); err != nil {
			s.ctx.Log.Errorln("[events] could not encode reply:", err)
			s.publish(msg.Reply(proto.InternalError(err)))
			return
		}

		s.publish(msg.Reply(reply))
	}
}

func (s *Service) publish(msg proto.Message) {
	if err := s.proto.Publish(msg); err != nil {
		s.ctx.Log.Errorln(err)
	}
}
