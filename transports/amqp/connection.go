// Copyright (C) 2019 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package amqp

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/transports/sfproto"
	"github.com/streadway/amqp"
)

type connection struct {
	conn     *connpair
	in       *amqp.Channel
	out      *amqp.Channel
	messages chan sarif.Message
	queue    amqp.Queue
	topics   []string
}

func Dial(cfg *sfproto.NetConfig) (sarif.Connection, error) {
	c, err := internalDial(cfg)
	if err != nil {
		return nil, err
	}

	return newConnection(c)
}

func newConnection(conn *connpair) (sarif.Connection, error) {
	aconn := &connection{conn: conn}
	aconn.messages = make(chan sarif.Message, 10)

	if err := aconn.setupOutgoing(); err != nil {
		return nil, err
	}
	if err := aconn.setupIncoming(); err != nil {
		return nil, err
	}

	go func() {
		ch := conn.out.NotifyReconnect(make(chan *amqp.Connection))
		for _ = range ch {
			aconn.setupOutgoing()
		}
	}()

	go func() {
		ch := conn.in.NotifyReconnect(make(chan *amqp.Connection))
		for _ = range ch {
			aconn.setupIncoming()
		}
	}()

	return aconn, nil
}

func (c *connection) setupOutgoing() (err error) {
	c.out, err = c.conn.out.conn.Channel()
	if err != nil {
		return err
	}

	err = c.out.ExchangeDeclare(
		"sarif",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	return err
}

func (c *connection) setupIncoming() (err error) {
	c.in, err = c.conn.in.conn.Channel()
	if err != nil {
		return err
	}

	c.queue, err = c.in.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for _, topic := range c.topics {
		if err := c.in.QueueBind(c.queue.Name, topic, "sarif", false, nil); err != nil {
			return err
		}
	}

	msgs, err := c.in.Consume(
		c.queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for amsg := range msgs {
			var msg sarif.Message
			err := json.Unmarshal(amsg.Body, &msg)
			if err != nil {
				continue
			}
			c.messages <- msg
		}
	}()

	return err
}

func (c *connection) Publish(msg sarif.Message) error {
	if err := msg.IsValid(); err != nil {
		return err
	}
	if c.conn == nil {
		err := errors.New("not connected")
		return err
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	topic := getTopic(msg.Action, msg.Destination)

	return c.out.Publish(
		"sarif",
		topic,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (c *connection) Subscribe(src, action, dest string) error {
	if dest == "self" {
		dest = src
	}
	topic := strings.TrimLeft(getTopic(action, dest)+".#", ".")
	if err := c.in.QueueBind(c.queue.Name, topic, "sarif", false, nil); err != nil {
		return err
	}
	c.topics = append(c.topics, topic)

	return nil
}

func (c *connection) Close() error {
	if c.conn == nil {
		return nil
	}
	if err := c.in.Close(); err != nil {
		return err
	}
	if err := c.out.Close(); err != nil {
		return err
	}
	c.conn = nil
	c.in = nil
	c.out = nil
	return nil
}

func (c *connection) Consume() (<-chan sarif.Message, error) {
	return (<-chan sarif.Message)(c.messages), nil
}
