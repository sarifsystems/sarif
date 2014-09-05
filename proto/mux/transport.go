// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mux

import "github.com/xconstruct/stark/proto"

type TransportMux struct {
	publisher proto.Publisher
	endpoints []*Endpoint
}

type Endpoint struct {
	mux     *TransportMux
	handler proto.Handler
	subs    []Subscription
}

func (e *Endpoint) Publish(msg proto.Message) error {
	if msg.Action == "proto/sub" {
		e.subs = append(e.subs, Subscription{
			Action: msg.PayloadGetString("action"),
			Device: msg.PayloadGetString("device"),
		})
	}
	return e.mux.publisher(msg)
}

func (e *Endpoint) RegisterHandler(h proto.Handler) {
	e.handler = h
}

func NewTransportMux() *TransportMux {
	m := &TransportMux{
		nil,
		make([]*Endpoint, 0),
	}
	return m
}

func (m *TransportMux) Handle(msg proto.Message) {
	for _, e := range m.endpoints {
		if e.handler == nil {
			continue
		}
		for _, s := range e.subs {
			if s.Matches(msg) {
				e.handler(msg)
				break
			}
		}
	}
}

func (m *TransportMux) RegisterPublisher(p proto.Publisher) {
	m.publisher = p
}

func (m *TransportMux) NewEndpoint() *Endpoint {
	e := &Endpoint{m, nil, make([]Subscription, 0)}
	m.endpoints = append(m.endpoints, e)
	return e
}
