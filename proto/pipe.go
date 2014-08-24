package proto

import (
	"errors"
)

var (
	ErrClosed = errors.New("The other pipe endpoint is closed.")
)

type PipeEndpoint struct {
	other   *PipeEndpoint
	closed  bool
	handler Handler
}

func NewPipe() (a, b *PipeEndpoint) {
	a = &PipeEndpoint{}
	b = &PipeEndpoint{}
	a.other = b
	b.other = a
	return a, b
}

func (t *PipeEndpoint) Publish(msg Message) error {
	if t.other == nil || t.other.closed {
		if t.other.closed {
			t.other = nil
		}
		return ErrClosed
	}

	if t.other.handler != nil {
		t.other.handler(msg)
	}
	return nil
}

func (t *PipeEndpoint) RegisterHandler(h Handler) {
	t.handler = h
}
