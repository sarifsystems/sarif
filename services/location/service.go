// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package location

import (
	"database/sql"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/proto/mux"
)

var Module = core.Module{
	Name:        "location",
	Version:     "1.0",
	NewInstance: nil,
}

func init() {
	core.RegisterModule(Module)
}

type Service struct {
	DB    Database
	ctx   *core.Context
	proto *proto.Client
}

func NewService(ctx *core.Context) (*Service, error) {
	db := ctx.Database

	s := &Service{
		&sqlDatabase{db.Driver(), db.DB},
		ctx,
		nil,
	}
	return s, nil
}

func NewInstance(ctx *core.Context) (core.ModuleInstance, error) {
	s, err := NewService(ctx)
	return s, err
}

func (s *Service) Enable() error {
	if err := s.DB.Setup(); err != nil {
		return err
	}

	mux := mux.New()
	mux.RegisterHandler("location/update", "", s.handleLocationUpdate)
	mux.RegisterHandler("location/last", "", s.handleLocationLast)

	s.proto = proto.NewClient("location", s.ctx.Proto)
	s.proto.RegisterHandler(mux.Handle)
	return s.proto.SubscribeGlobal("location/update")
}

func (s *Service) Disable() error {
	return nil
}

func (s *Service) handleLocationUpdate(msg proto.Message) {
	loc := Location{}
	if err := msg.DecodePayload(&loc); err != nil {
		s.ctx.Log.Warnln(err)
		reply := msg.Reply(proto.BadRequest(err))
		s.ctx.Must(s.proto.Publish(reply))
		return
	}
	s.ctx.Log.Debugln(loc)

	if err := s.DB.Store(loc); err != nil {
		s.ctx.Log.Errorln(err)
	}
}

type locationLastMessage struct {
	Address   string
	Bounds    []float64
	Latitude  float64
	Longitude float64
	Accuracy  float64
}

func (s *Service) handleLocationLast(msg proto.Message) {
	var pl locationLastMessage
	if err := msg.DecodePayload(&pl); err != nil {
		s.ctx.Log.Warnln(err)
		reply := msg.Reply(proto.BadRequest(err))
		if err := s.proto.Publish(reply); err != nil {
			s.ctx.Log.Errorln(err)
		}
		return
	}

	s.ctx.Log.Debugln(pl)
	var reply proto.Message
	if len(pl.Bounds) == 4 {
		loc, err := s.DB.GetLastLocationInBounds(
			pl.Bounds[0], pl.Bounds[1], pl.Bounds[2], pl.Bounds[3])

		if err == sql.ErrNoRows {
			reply = proto.Message{
				Action: "location/notfound",
				Payload: map[string]interface{}{
					"text": "Requested location could not be found",
				},
			}
		} else if err != nil {
			reply = proto.InternalError(err)
		} else {
			reply = proto.Message{
				Action: "location/found",
			}
			if err := reply.EncodePayload(loc); err != nil {
				s.ctx.Log.Errorln(err)
			}
		}
	} else {
		reply = msg.Reply(proto.BadRequest(nil))
	}

	if reply.Action != "" {
		reply = msg.Reply(reply)
		if err := s.proto.Publish(reply); err != nil {
			s.ctx.Log.Errorln(err)
		}
	}
}
