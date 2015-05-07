// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package store

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "store",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	DB     *gorm.DB
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	Store Store
	Log   proto.Logger
	*proto.Client
}

func NewService(deps *Dependencies) *Service {
	return &Service{
		Store:  &sqlStore{deps.DB},
		Log:    deps.Log,
		Client: deps.Client,
	}
}

func (s *Service) Enable() error {
	if err := s.Store.Setup(); err != nil {
		return err
	}
	s.Subscribe("store/put", "", s.handlePut)
	s.Subscribe("store/get", "", s.handleGet)
	s.Subscribe("store/del", "", s.handleDel)
	s.Subscribe("store/scan", "", s.handleScan)
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
		s.Log.Warnln("[store] received bad payload:", err)
		s.ReplyBadRequest(msg, err)
		return
	}
	doc, err := s.Store.Put(Document{
		Key:   key,
		Value: []byte(value),
	})
	if err != nil {
		s.Log.Errorln("[store] could not store:", err)
		s.ReplyInternalError(msg, err)
		return
	}

	replydoc := doc
	replydoc.Value = nil
	reply := proto.CreateMessage("store/updated/"+doc.Key, replydoc)
	s.Reply(msg, reply)

	pub := proto.CreateMessage("store/updated/"+doc.Key, doc.Value)
	pub.Text = "Document " + doc.Key + "."
	s.Publish(pub)
}

func (s *Service) handleGet(msg proto.Message) {
	key := ""
	if strings.HasPrefix(msg.Action, "store/get/") {
		key = strings.TrimPrefix(msg.Action, "store/get/")
	}
	if key == "" {
		s.ReplyBadRequest(msg, errors.New("No key specified."))
		return
	}

	doc, err := s.Store.Get(key)
	if err == ErrNoResult {
		s.Publish(msg.Reply(proto.Message{
			Action: "err/notfound",
			Text:   "Document " + key + " not found.",
		}))
		return
	} else if err != nil {
		s.Log.Errorln("[store] could not retrieve:", err)
		s.ReplyInternalError(msg, err)
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
	s.Reply(msg, reply)
}

func (s *Service) handleDel(msg proto.Message) {
	key := ""
	if strings.HasPrefix(msg.Action, "store/del/") {
		key = strings.TrimPrefix(msg.Action, "store/del/")
	}
	if key == "" {
		s.ReplyBadRequest(msg, errors.New("No key specified."))
		return
	}
	if err := s.Store.Del(key); err != nil {
		s.Log.Errorln("[store] could not delete:", err)
		s.ReplyInternalError(msg, err)
		return
	}

	s.Reply(msg, proto.CreateMessage("store/deleted/"+key, nil))
	s.Publish(proto.CreateMessage("store/deleted/"+key, nil))
}

type scanMessage struct {
	Prefix string `json:"prefix"`
	Start  string `json:"start"`
	End    string `json:"end"`
}

func (s *Service) handleScan(msg proto.Message) {
	var p scanMessage
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	if p.Prefix == "" && strings.HasPrefix(msg.Action, "store/scan/") {
		p.Prefix = strings.TrimPrefix(msg.Action, "store/scan/")
	}
	keys, err := s.Store.Scan(p.Prefix, p.Start, p.End)
	if err != nil {
		s.Log.Errorln("[store] could not scan:", err)
		s.ReplyInternalError(msg, err)
		return
	}
	s.Reply(msg, proto.CreateMessage("store/scanned", keys))
}
