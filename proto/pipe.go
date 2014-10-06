// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"errors"
)

var (
	ErrClosed = errors.New("The other end is closed.")
)

type pipeConn struct {
	other   *pipeConn
	closed  bool
	handler Handler
}

func NewPipe() (a, b Conn) {
	ac := &pipeConn{}
	bc := &pipeConn{}
	ac.other = bc
	bc.other = ac
	return ac, bc
}

func (t *pipeConn) Publish(msg Message) error {
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

func (t *pipeConn) RegisterHandler(h Handler) {
	t.handler = h
}
