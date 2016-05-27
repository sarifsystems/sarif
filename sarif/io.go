// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sarif

import (
	"encoding/json"
	"io"
)

type byteConn struct {
	conn io.ReadWriteCloser
	enc  *json.Encoder
	dec  *json.Decoder
}

func NewByteConn(conn io.ReadWriteCloser) Conn {
	t := &byteConn{
		conn,
		json.NewEncoder(conn),
		json.NewDecoder(conn),
	}
	return t
}

func (t *byteConn) Write(msg Message) error {
	if err := msg.IsValid(); err != nil {
		return err
	}
	return t.enc.Encode(msg)
}

func (t *byteConn) Read() (Message, error) {
	var msg Message
	if err := t.dec.Decode(&msg); err != nil {
		return msg, err
	}
	return msg, msg.IsValid()
}

func (t *byteConn) Close() error {
	return t.conn.Close()
}
