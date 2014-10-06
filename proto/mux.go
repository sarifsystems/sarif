// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

type Mux struct {
	publisher Publisher
	conns     []*muxConn
}

type muxConn struct {
	mux     *Mux
	handler Handler
	subs    []subscription
}

func (e *muxConn) Publish(msg Message) error {
	if msg.IsAction("proto/sub") {
		sub := subscription{}
		if err := msg.DecodePayload(&sub); err == nil {
			e.subs = append(e.subs, sub)
		}
	}
	return e.mux.publisher(msg)
}

func (e *muxConn) RegisterHandler(h Handler) {
	e.handler = h
}

func NewMux() *Mux {
	m := &Mux{
		nil,
		make([]*muxConn, 0),
	}
	return m
}

func (m *Mux) Handle(msg Message) {
	for _, e := range m.conns {
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

func (m *Mux) NewConn() Conn {
	e := &muxConn{m, nil, make([]subscription, 0)}
	m.conns = append(m.conns, e)
	return e
}
