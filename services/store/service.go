// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package store

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
)

var Module = core.Module{
	Name:        "store",
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
	Store Store
	ctx   *core.Context
	proto *proto.Client
}

func NewService(ctx *core.Context) (*Service, error) {
	s := &Service{
		Store: &sqlStore{ctx.Database.Driver(), ctx.Database.DB},
		ctx:   ctx,
		proto: proto.NewClient("store", ctx.Proto),
	}
	return s, nil
}

func (s *Service) Enable() error {
	if err := s.Store.Setup(); err != nil {
		return err
	}
	if err := s.proto.Subscribe("store/put", "", s.handlePut); err != nil {
		return err
	}
	if err := s.proto.Subscribe("store/get", "", s.handleGet); err != nil {
		return err
	}
	if err := s.proto.Subscribe("store/del", "", s.handleDel); err != nil {
		return err
	}
	return nil
}

func (s *Service) Disable() error { return nil }

func (s *Service) handlePut(msg proto.Message) {
	key := ""
	if strings.HasPrefix(msg.Action, "store/put/") {
		key = strings.TrimPrefix(msg.Action, "store/put/")
	}

	var value json.RawMessage
	if msg.Payload == nil && msg.Text != "" {
		v, _ := json.Marshal(msg.Text)
		value = json.RawMessage(v)
	}
	if err := msg.DecodePayload(&value); err != nil {
		s.ctx.Log.Warnln("[store] received bad payload:", err)
		s.proto.Publish(msg.Reply(proto.BadRequest(err)))
		return
	}
	doc, err := s.Store.Put(Document{
		Key:   key,
		Value: []byte(value),
	})
	if err != nil {
		s.ctx.Log.Errorln("[store] could not store:", err)
		s.proto.Publish(msg.Reply(proto.InternalError(err)))
		return
	}

	replydoc := doc
	replydoc.Value = nil
	reply := proto.CreateMessage("store/updated/"+doc.Key, replydoc)
	s.proto.Publish(msg.Reply(reply))

	pub := proto.CreateMessage("store/updated/"+doc.Key, doc.Value)
	pub.Text = "Document " + doc.Key + "."
	s.proto.Publish(pub)
}

func (s *Service) handleGet(msg proto.Message) {
	key := ""
	if strings.HasPrefix(msg.Action, "store/get/") {
		key = strings.TrimPrefix(msg.Action, "store/get/")
	}
	if key == "" {
		s.proto.Publish(msg.Reply(proto.BadRequest(errors.New("No key specified."))))
		return
	}

	doc, err := s.Store.Get(key)
	if err == ErrNoResult {
		s.proto.Publish(msg.Reply(proto.Message{
			Action: "err/notfound",
			Text:   "Document " + key + " not found.",
		}))
		return
	} else if err != nil {
		s.ctx.Log.Errorln("[store] could not retrieve:", err)
		s.proto.Publish(msg.Reply(proto.InternalError(err)))
		return
	}

	raw := json.RawMessage(doc.Value)
	reply := proto.CreateMessage("store/retrieved/"+doc.Key, &raw)
	reply.Text = "Document " + doc.Key + "."
	var val string
	if err := json.Unmarshal(doc.Value, &val); err == nil {
		if len(val) < 200 {
			reply.Text = "Document " + doc.Key + `: "` + val + `".`
		}
	}
	s.proto.Publish(msg.Reply(reply))
}

func (s *Service) handleDel(msg proto.Message) {
	key := ""
	if strings.HasPrefix(msg.Action, "store/del/") {
		key = strings.TrimPrefix(msg.Action, "store/del/")
	}
	if key == "" {
		s.proto.Publish(msg.Reply(proto.BadRequest(errors.New("No key specified."))))
		return
	}
	if err := s.Store.Del(key); err != nil {
		s.ctx.Log.Errorln("[store] could not delete:", err)
		s.proto.Publish(msg.Reply(proto.InternalError(err)))
		return
	}

	s.proto.Publish(msg.Reply(proto.CreateMessage("store/deleted/"+key, nil)))
	s.proto.Publish(proto.CreateMessage("store/deleted/"+key, nil))
}
