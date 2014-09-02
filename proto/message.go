// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"encoding/json"
	"errors"
)

const VERSION = "0.3"

type Message struct {
	Version     string                 `json:"v"`
	Id          string                 `json:"id"`
	Action      string                 `json:"action"`
	Source      string                 `json:"src"`
	Destination string                 `json:"dst,omitempty"`
	Payload     map[string]interface{} `json:"p,omitempty"`
	CorrId      string                 `json:"corr,omitempty"`
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

func (m Message) PayloadGetString(key string) string {
	if val, ok := m.Payload[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (m Message) DecodePayload(v interface{}) error {
	raw, err := json.Marshal(m.Payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, v)
}

type stringer interface {
	String() string
}

func (m *Message) EncodePayload(v interface{}) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	m.Payload = nil
	if err := json.Unmarshal(raw, &m.Payload); err != nil {
		return err
	}
	if _, ok := m.Payload["text"]; !ok {
		if s, ok := v.(stringer); ok {
			m.Payload["text"] = s.String()
		}
	}
	return nil
}
