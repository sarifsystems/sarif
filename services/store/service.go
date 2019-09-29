// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service store provides a key-value store to the sarif network.
package store

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/sarifsystems/sarif/pkg/mapq"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
)

var Module = &services.Module{
	Name:        "store",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	Driver string
	Path   string
}

type Dependencies struct {
	Config services.Config
	Client sarif.Client
}

type Service struct {
	Config services.Config
	Cfg    Config
	Store  Store
	sarif.Client
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Config: deps.Config,
		Client: deps.Client,
	}
	return s
}

func (s *Service) Enable() (err error) {
	if _, ok := drivers["bolt"]; ok {
		s.Cfg.Driver = "bolt"
		s.Cfg.Path = s.Config.Dir() + "/" + "store.bolt.db"
	} else if _, ok := drivers["memory"]; ok {
		s.Cfg.Driver = "mem"
	}
	s.Config.Get(&s.Cfg)

	if s.Cfg.Driver == "" {
		return errors.New("Store: no database driver set in config!")
	}
	drv, ok := drivers[s.Cfg.Driver]
	if !ok {
		return errors.New("Store: no driver with name '" + s.Cfg.Driver + "' found!")
	}

	if s.Store, err = drv.Open(s.Cfg.Path); err != nil {
		return err
	}

	s.Subscribe("store/put", "", s.handlePut)
	s.Subscribe("store/get", "", s.handleGet)
	s.Subscribe("store/del", "", s.handleDel)
	s.Subscribe("store/scan", "", s.handleScan)
	s.Subscribe("store/batch", "", s.handleBatch)
	return nil
}

func (s *Service) Disable() error { return nil }

func parseAction(prefix, action string) (col, key string) {
	if !strings.HasPrefix(action, prefix) {
		return "", ""
	}
	colkey := strings.TrimPrefix(action, prefix)
	parts := strings.SplitN(colkey, "/", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func (s *Service) handlePut(msg sarif.Message) {
	collection, key := parseAction("store/put/", msg.Action)
	if collection == "" {
		s.ReplyBadRequest(msg, errors.New("No collection specified."))
		return
	}

	if len(msg.Payload.Raw) == 0 && msg.Text != "" {
		v, _ := json.Marshal(msg.Text)
		msg.Payload.Raw = json.RawMessage(v)
	}
	var p interface{}
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	// TODO: maybe a JSON payload consistency check
	doc, err := s.Store.Put(&Document{
		Collection: collection,
		Key:        key,
		Value:      msg.Payload.Raw,
	})
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	doc.Value = nil
	reply := sarif.CreateMessage("store/updated/"+doc.Collection+"/"+doc.Key, doc)
	s.Reply(msg, reply)

	pub := sarif.CreateMessage("store/updated/"+doc.Collection+"/"+doc.Key, doc)
	s.Publish(pub)
}

func (s *Service) handleGet(msg sarif.Message) {
	collection, key := parseAction("store/get/", msg.Action)
	if collection == "" || key == "" {
		s.ReplyBadRequest(msg, errors.New("No collection or key specified."))
		return
	}

	doc, err := s.Store.Get(collection, key)
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	} else if doc == nil {
		s.Reply(msg, sarif.Message{
			Action: "err/notfound",
			Text:   "Document " + collection + "/" + key + " not found.",
		})
		return
	}

	raw := json.RawMessage(doc.Value)
	reply := sarif.CreateMessage("store/retrieved/"+doc.Collection+"/"+doc.Key, nil)
	reply.Payload.Encode(&raw)
	s.Reply(msg, reply)
}

func (s *Service) handleDel(msg sarif.Message) {
	collection, key := parseAction("store/del/", msg.Action)
	if collection == "" || key == "" {
		s.ReplyBadRequest(msg, errors.New("No collection or key specified."))
		return
	}
	if err := s.Store.Del(collection, key); err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	s.Reply(msg, sarif.CreateMessage("store/deleted/"+collection+"/"+key, nil))
	s.Publish(sarif.CreateMessage("store/deleted/"+collection+"/"+key, nil))
}

type BatchCommand struct {
	Type  string          `json:"type"`
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value,omitempty"`
}

func (s *Service) handleBatch(msg sarif.Message) {
	cmds := make([]BatchCommand, 0)
	if err := msg.DecodePayload(&cmds); err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	results := make([]interface{}, len(cmds))
	for i, cmd := range cmds {
		collection, key := parseAction("", cmd.Key)
		switch cmd.Type {
		case "put":
			doc, err := s.Store.Put(&Document{
				Collection: collection,
				Key:        key,
				Value:      []byte(cmd.Value),
			})
			if err != nil {
				s.ReplyInternalError(msg, err)
				return
			}
			results[i] = doc
		case "get":
			doc, err := s.Store.Get(collection, key)
			if err != nil {
				s.ReplyInternalError(msg, err)
				return
			}
			results[i] = doc
		case "del":
			if err := s.Store.Del(collection, key); err != nil {
				s.ReplyInternalError(msg, err)
				return
			}
			results[i] = true
		case "scan":
			var p scanMessage
			if err := json.Unmarshal(cmd.Value, &p); err != nil {
				s.ReplyInternalError(msg, err)
				return
			}
			got, err := s.doScan(collection, p)
			if err != nil {
				s.ReplyInternalError(msg, err)
				return
			}
			results[i] = got
		}
	}

	s.Reply(msg, sarif.CreateMessage("store/batched/", results))
}

type scanMessage struct {
	Prefix  string `json:"prefix"`
	Start   string `json:"start"`
	End     string `json:"end"`
	Only    string `json:"only"`
	Limit   int    `json:"limit"`
	Reverse bool   `json:"reverse"`

	Filter map[string]interface{} `json:"filter"`
}

func (s *Service) handleScan(msg sarif.Message) {
	collection, prefix := parseAction("store/scan/", msg.Action)
	if collection == "" {
		s.ReplyBadRequest(msg, errors.New("No collection specified."))
		return
	}

	var p scanMessage
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	if p.Start == "" && p.End == "" {
		if p.Prefix == "" {
			p.Prefix = prefix
		}
	}

	got, err := s.doScan(collection, p)
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	s.Reply(msg, sarif.CreateMessage("store/scanned/"+collection, got))
}

type DocsPayload struct {
	Keys   []string           `json:"keys"`
	Values []*json.RawMessage `json:"values"`
}

func (s *Service) doScan(collection string, p scanMessage) (interface{}, error) {
	if p.Prefix != "" {
		start, end := clampPrefix(p.Prefix)
		if p.Start == "" {
			p.Start = start
		}
		if p.End == "" {
			p.End = end
		}
	}
	if p.Limit == 0 {
		p.Limit = 100
	}

	cursor, err := s.Store.Scan(collection, p.Start, p.End, p.Reverse)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	keys := make([]string, 0)
	values := make([]*json.RawMessage, 0)
	for doc := cursor.Next(); doc != nil && p.Limit > 0; doc = cursor.Next() {
		if p.Filter != nil {
			var object map[string]interface{}
			if err := json.Unmarshal(doc.Value, &object); err != nil {
				continue
			}
			if !mapq.M(object).Matches(mapq.Filter(p.Filter)) {
				continue
			}
		}

		p.Limit--
		if p.Only == "" || p.Only == "keys" {
			keys = append(keys, doc.Key)
		}
		if p.Only == "" || p.Only == "values" {
			raw := json.RawMessage(doc.Value)
			values = append(values, &raw)
		}
	}

	if p.Only == "keys" {
		return keys, nil
	} else if p.Only == "values" {
		return values, nil
	} else {
		return DocsPayload{keys, values}, nil
	}
}

func clampPrefix(prefix string) (start, end string) {
	return prefix, prefix + "~~~~~"
}
