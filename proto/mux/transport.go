package mux

import "github.com/xconstruct/stark/proto"

type TransportMux struct {
	publisher proto.Publisher
	endpoints []*Endpoint
	*Mux
}

type Endpoint struct {
	mux     *TransportMux
	handler proto.Handler
}

func (e *Endpoint) Publish(msg proto.Message) error {
	if msg.Action == "proto/sub" {
		action := msg.PayloadGetString("action")
		device := msg.PayloadGetString("device")
		e.mux.RegisterHandler(action, device, e.handle)
	}
	return e.mux.publisher(msg)
}

func (e *Endpoint) handle(msg proto.Message) {
	e.handler(msg)
}

func (e *Endpoint) RegisterHandler(h proto.Handler) {
	e.handler = h
}

func NewTransportMux() *TransportMux {
	return &TransportMux{
		nil,
		make([]*Endpoint, 0),
		New(),
	}
}

func (m *TransportMux) RegisterPublisher(p proto.Publisher) {
	m.publisher = p
}

func (m *TransportMux) NewEndpoint() *Endpoint {
	e := &Endpoint{m, nil}
	m.endpoints = append(m.endpoints, e)
	return e
}
