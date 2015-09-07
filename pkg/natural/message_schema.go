// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import "github.com/xconstruct/stark/proto"

type MessageSchema struct {
	Action string
	Fields map[string]string
}

func (r *MessageSchema) Apply(vars []*Var) proto.Message {
	msg := proto.Message{}
	msg.Action = r.Action
	pl := make(map[string]string)
	for _, v := range vars {
		switch v.Name {
		case "_action":
			msg.Action = v.Value
		case "text":
			msg.Text = v.Value
		case "to":
			fallthrough
		case "that":
			if msg.Text == "" {
				msg.Text = v.Value
			}
		default:
			if _, ok := r.Fields[v.Name]; ok {
				pl[v.Name] = v.Value
			}
		}
	}
	if len(vars) > 0 {
		msg.EncodePayload(&pl)
	}
	return msg
}

type MessageSchemaStore struct {
	Messages map[string]*MessageSchema
}

func NewMessageSchemaStore() *MessageSchemaStore {
	return &MessageSchemaStore{
		make(map[string]*MessageSchema),
	}
}

func (s *MessageSchemaStore) Get(action string) *MessageSchema {
	return s.Messages[action]
}

func (s *MessageSchemaStore) Set(schema *MessageSchema) {
	s.Messages[schema.Action] = schema
}

func (s *MessageSchemaStore) Add(schema *MessageSchema) {
	ex := s.Messages[schema.Action]
	if ex == nil {
		s.Set(schema)
		return
	}

	for name, typ := range schema.Fields {
		ex.Fields[name] = typ
	}
}

func (s *MessageSchemaStore) AddMessage(msg *proto.Message) {
	schema := &MessageSchema{
		Action: msg.Action,
		Fields: make(map[string]string),
	}
	p := make(map[string]interface{})
	msg.DecodePayload(&p)
	for v, _ := range p {
		schema.Fields[v] = "text"
	}

	s.Add(schema)
}

func (s *MessageSchemaStore) AddDataSet(set DataSet) {
	for _, data := range set {
		schema := &MessageSchema{
			Action: data.Action,
			Fields: make(map[string]string),
		}
		for _, v := range data.Vars {
			schema.Fields[v.Name] = v.Type
		}
		s.Add(schema)
	}
}
