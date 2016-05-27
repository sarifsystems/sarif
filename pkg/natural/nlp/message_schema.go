// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package nlp

import (
	"sort"

	"github.com/sarifsystems/sarif/sarif"
)

type MessageSchema struct {
	Action string
	Fields map[string]string
}

type byWeight []*Var

func (a byWeight) Len() int           { return len(a) }
func (a byWeight) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byWeight) Less(i, j int) bool { return a[i].Weight > a[j].Weight }

func (r *MessageSchema) Apply(vars []*Var) sarif.Message {
	msg := sarif.Message{}
	msg.Action = r.Action
	pl := make(map[string]string)
	sort.Sort(sort.Reverse(byWeight(vars)))

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

func (s *MessageSchemaStore) AddMessage(msg *sarif.Message) {
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
