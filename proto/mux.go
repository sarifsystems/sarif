// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

type Mux struct {
	publisher Publisher
	endpoints []*MuxEndpoint
}

type MuxEndpoint struct {
	mux     *Mux
	handler Handler
	subs    []Subscription
}

func (e *MuxEndpoint) Publish(msg Message) error {
	if msg.IsAction("proto/sub") {
		e.subs = append(e.subs, Subscription{
			Action: msg.PayloadGetString("action"),
			Device: msg.PayloadGetString("device"),
		})
	}
	return e.mux.publisher(msg)
}

func (e *MuxEndpoint) RegisterHandler(h Handler) {
	e.handler = h
}

func NewMux() *Mux {
	m := &Mux{
		nil,
		make([]*MuxEndpoint, 0),
	}
	return m
}

func (m *Mux) Handle(msg Message) {
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

func (m *Mux) RegisterPublisher(p Publisher) {
	m.publisher = p
}

func (m *Mux) NewEndpoint() *MuxEndpoint {
	e := &MuxEndpoint{m, nil, make([]Subscription, 0)}
	m.endpoints = append(m.endpoints, e)
	return e
}
