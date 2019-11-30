// Copyright (C) 2019 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package amqp

import (
	"log"
	"time"

	"github.com/sarifsystems/sarif/transports/sfproto"
	"github.com/streadway/amqp"
)

type amqpConn struct {
	cfg       *sfproto.NetConfig
	conn      *amqp.Connection
	listeners []chan *amqp.Connection
}

func (c *amqpConn) Dial() (err error) {
	c.conn, err = amqp.Dial(c.cfg.Address)
	if err != nil {
		return err
	}

	go c.reconnector()

	return nil
}

func (c *amqpConn) NotifyReconnect(listener chan *amqp.Connection) chan *amqp.Connection {
	c.listeners = append(c.listeners, listener)

	return listener
}

func (c *amqpConn) reconnector() {
	reason, ok := <-c.conn.NotifyClose(make(chan *amqp.Error))
	if !ok {
		return
	}

	log.Printf("connection closed, reason: %v", reason)
	for {
		time.Sleep(3 * time.Second)
		err := c.Dial()
		if err == nil {
			log.Printf("connection reconnected")
			for _, l := range c.listeners {
				l <- c.conn
			}
			return
		}

		log.Printf("connection reconnect failed, reason: %v", err)
	}
}
