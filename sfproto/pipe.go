// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sfproto

import (
	"errors"

	"github.com/sarifsystems/sarif/sarif"
)

var (
	ErrClosed = errors.New("The other end is closed.")
)

type pipeConn struct {
	name     string
	other    *pipeConn
	messages chan sarif.Message
}

func NewPipe() (a, b Conn) {
	ac := &pipeConn{}
	bc := &pipeConn{}
	ac.messages = make(chan sarif.Message, 10)
	bc.messages = make(chan sarif.Message, 10)
	ac.other = bc
	bc.other = ac

	id := sarif.GenerateId()
	ac.name = "Pipe-" + id + "-a"
	bc.name = "Pipe-" + id + "-b"
	return ac, bc
}

func (t *pipeConn) Write(msg sarif.Message) error {
	if err := msg.IsValid(); err != nil {
		return err
	}
	if t.other == nil {
		return ErrClosed
	}
	t.other.messages <- msg
	return nil
}

func (t *pipeConn) Read() (sarif.Message, error) {
	msg, ok := <-t.messages
	if !ok {
		return msg, ErrClosed
	}
	return msg, nil
}

func (t *pipeConn) Close() error {
	if t.other == nil {
		return nil
	}

	o := t.other
	t.other = nil
	close(t.messages)
	return o.Close()
}

func (t *pipeConn) String() string {
	return t.name
}
