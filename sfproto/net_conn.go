// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sfproto

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"time"

	"github.com/sarifsystems/sarif/sarif"
)

type netConn struct {
	conn     net.Conn
	enc      *json.Encoder
	dec      *json.Decoder
	verified bool
}

func newNetConn(conn net.Conn) *netConn {
	return &netConn{
		conn,
		json.NewEncoder(conn),
		json.NewDecoder(conn),
		false,
	}
}

func (c *netConn) Write(msg sarif.Message) error {
	if err := msg.IsValid(); err != nil {
		return err
	}
	c.conn.SetWriteDeadline(time.Now().Add(5 * time.Minute))
	return c.enc.Encode(msg)
}

func (c *netConn) KeepaliveLoop(ka time.Duration) error {
	for {
		time.Sleep(ka)
		c.conn.SetWriteDeadline(time.Now().Add(3 * ka))
		if _, err := c.conn.Write([]byte(" ")); err != nil {
			return err
		}
	}
}

func (c *netConn) Read() (sarif.Message, error) {
	var msg sarif.Message
	c.conn.SetReadDeadline(time.Now().Add(time.Hour))
	if err := c.dec.Decode(&msg); err != nil {
		return msg, err
	}
	return msg, msg.IsValid()
}

func (c *netConn) Close() error {
	return c.conn.Close()
}

func (c *netConn) IsVerified() bool {
	if tc, ok := c.conn.(*tls.Conn); ok {
		if len(tc.ConnectionState().VerifiedChains) > 0 {
			return true
		}
	}

	return false
}
