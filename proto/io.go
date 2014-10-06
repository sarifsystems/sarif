// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"encoding/json"
	"io"
)

type ByteConn struct {
	enc     *json.Encoder
	dec     *json.Decoder
	handler Handler
}

func NewByteConn(conn io.ReadWriter) *ByteConn {
	t := &ByteConn{
		json.NewEncoder(conn),
		json.NewDecoder(conn),
		nil,
	}
	return t
}

func (t *ByteConn) Publish(msg Message) error {
	return t.enc.Encode(msg)
}

func (t *ByteConn) RegisterHandler(h Handler) {
	t.handler = h
}

func (t *ByteConn) Listen() error {
	for {
		var msg Message
		if err := t.dec.Decode(&msg); err != nil {
			return err
		}

		if t.handler != nil {
			t.handler(msg)
		}
	}
}
