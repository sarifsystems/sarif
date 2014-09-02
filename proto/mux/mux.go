// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mux

import (
	"strings"

	"github.com/xconstruct/stark/proto"
)

type Subscription struct {
	Action  string
	Device  string
	Handler proto.Handler
}

func (s Subscription) Matches(msg proto.Message) bool {
	if msg.Destination != s.Device {
		return false
	}
	if s.Action != "" && !strings.HasPrefix(msg.Action+"/", s.Action+"/") {
		return false
	}
	return true
}

type Mux struct {
	concurrent    bool
	subscriptions []Subscription
}

func New() *Mux {
	return &Mux{
		false,
		make([]Subscription, 0),
	}
}

func (m *Mux) RegisterHandler(action, device string, h proto.Handler) {
	m.subscriptions = append(m.subscriptions, Subscription{
		action,
		device,
		h,
	})
}

func (m *Mux) Handle(msg proto.Message) {
	for _, s := range m.subscriptions {
		if s.Matches(msg) {
			if m.concurrent {
				go s.Handler(msg)
			} else {
				s.Handler(msg)
			}
		}
	}
}
