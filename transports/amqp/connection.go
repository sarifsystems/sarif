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
	conn  *connpair
	in    *amqp.Channel
	out   *amqp.Channel
	queue amqp.Queue
}

type connpair struct {
	in  *amqp.Connection
	out *amqp.Connection
}

func internalDial(cfg *sfproto.NetConfig) (*connpair, error) {
	in, err := amqp.Dial(cfg.Address)
	if err != nil {
		return nil, err
	}

	out, err := amqp.Dial(cfg.Address)
	if err != nil {
		return nil, err
	}

	return &connpair{in, out}, nil
}

func Dial(cfg *sfproto.NetConfig) (sarif.Connection, error) {
	c, err := internalDial(cfg)
	if err != nil {
		return nil, err
	}

	return newConnection(c)
}

func newConnection(conn *connpair) (sarif.Connection, error) {
	out, err := conn.out.Channel()
	if err != nil {
		return nil, err
	}

	err = out.ExchangeDeclare(
		"sarif",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	in, err := conn.in.Channel()
	if err != nil {
		return nil, err
	}

	queue, err := in.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &connection{conn, in, out, queue}, nil
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
	ch := make(chan sarif.Message, 10)

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
		return nil, err
	}

	go func() {
		for amsg := range msgs {
			var msg sarif.Message
			err := json.Unmarshal(amsg.Body, &msg)
			if err != nil {
				close(ch)
				c.conn = nil
				return
			}
			ch <- msg
		}
	}()

	return (<-chan sarif.Message)(ch), nil
}
