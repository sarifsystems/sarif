// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package proto implements the stark protocol, including client and broker.
package proto

import (
	"errors"
	"strings"
)

const VERSION = "0.5"

type Message struct {
	Version     string  `json:"stark,omitempty"`
	Id          string  `json:"id,omitempty"`
	Action      string  `json:"action,omitempty"`
	Source      string  `json:"src,omitempty"`
	Destination string  `json:"dst,omitempty"`
	Payload     Partial `json:"p,omitempty"`
	CorrId      string  `json:"corr,omitempty"`
	Text        string  `json:"text,omitempty"`
}

func CreateMessage(action string, payload interface{}) Message {
	msg := Message{
		Version: VERSION,
		Id:      GenerateId(),
		Action:  action,
	}
	if err := msg.EncodePayload(payload); err != nil {
		panic(err)
	}

	if s, ok := payload.(interface {
		Text() string
	}); ok {
		msg.Text = s.Text()
	} else if s, ok := payload.(interface {
		String() string
	}); ok {
		msg.Text = s.String()
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
		if m.CorrId = orig.CorrId; m.CorrId == "" {
			m.CorrId = orig.Id
		}
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
	return m.Payload.Decode(v)
}

func (m *Message) EncodePayload(v interface{}) error {
	return m.Payload.Encode(v)
}
