// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"encoding/json"
	"errors"
	"strings"
)

const VERSION = "0.4"

type Message struct {
	Version     string           `json:"v"`
	Id          string           `json:"id"`
	Action      string           `json:"action"`
	Source      string           `json:"src"`
	Destination string           `json:"dst,omitempty"`
	Payload     *json.RawMessage `json:"p,omitempty"`
	CorrId      string           `json:"corr,omitempty"`
	Text        string           `json:"text,omitempty"`
}

func (m Message) Encode() ([]byte, error) {
	if err := m.IsValid(); err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

func DecodeMessage(raw []byte) (Message, error) {
	m := Message{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return m, err
	}
	if err := m.IsValid(); err != nil {
		return m, err
	}
	return m, nil
}

func CreateMessage(action string, payload interface{}) Message {
	msg := Message{
		Id:     GenerateId(),
		Action: action,
	}
	if err := msg.EncodePayload(payload); err != nil {
		panic(err)
	}
	return msg
}

func (m Message) IsValid() error {
	if m.Version == "" {
		return errors.New("Invalid stark message: missing version")
	}
	if m.Id == "" {
		return errors.New("Invalid stark message: missing id")
	}
	if m.Action == "" {
		return errors.New("Invalid stark message: missing action")
	}
	if m.Source == "" {
		return errors.New("Invalid stark message: missing source")
	}
	return nil
}

func (orig Message) Reply(m Message) Message {
	if m.CorrId == "" {
		m.CorrId = orig.Id
	}
	if m.Destination == "" {
		m.Destination = orig.Source
	}
	return m
}

func (m Message) IsAction(action string) bool {
	if action != "" && !strings.HasPrefix(m.Action+"/", action+"/") {
		return false
	}
	return true
}

func (m Message) DecodePayload(v interface{}) error {
	if m.Payload == nil {
		return nil
	}
	return json.Unmarshal(*m.Payload, v)
}

type stringer interface {
	String() string
}

func (m *Message) EncodePayload(v interface{}) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	rawjson := json.RawMessage(raw)
	m.Payload = &rawjson
	if m.Text == "" {
		if s, ok := v.(stringer); ok {
			m.Text = s.String()
		}
	}
	return nil
}

func (m Message) Copy() Message {
	c := m
	if m.Payload != nil {
		if err := c.EncodePayload(m.Payload); err != nil {
			panic(err)
		}
	}
	return c
}
