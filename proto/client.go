// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"github.com/xconstruct/stark/log"
)

type Client struct {
	DeviceName string
	DeviceId   string
	endpoint   Endpoint
	handler    Handler
	log        log.Interface
}

func NewClient(deviceName string, e Endpoint) *Client {
	c := &Client{
		deviceName,
		deviceName + "-" + GenerateId(),
		e,
		nil,
		log.Default,
	}
	c.SetEndpoint(e)
	return c
}

func (c *Client) SetLogger(l log.Interface) {
	c.log = l
}

func (c *Client) SetEndpoint(e Endpoint) {
	c.endpoint = e
	e.RegisterHandler(c.handle)
}

func (c *Client) FillMessage(msg *Message) {
	if msg.Version == "" {
		msg.Version = VERSION
	}
	if msg.Id == "" {
		msg.Id = GenerateId()
	}
	if msg.Source == "" {
		msg.Source = c.DeviceId
	}
}

func (c *Client) Publish(msg Message) error {
	c.FillMessage(&msg)
	return c.endpoint.Publish(msg)
}

func (c *Client) handle(msg Message) {
	if msg.Action == "ping" {
		c.handlePing(msg)
	}
	c.handler(msg)
}

func (c *Client) handlePing(msg Message) {
	c.log.Debugf("%s got ping", c.DeviceId)
	err := c.Publish(msg.Reply(Message{
		Action: "ack",
		CorrId: msg.Id,
	}))
	if err != nil {
		c.log.Warnln(err)
	}
}

func (c *Client) RegisterHandler(h Handler) {
	c.handler = h
	c.SubscribeGlobal("ping")
}

func (c *Client) SubscribeGlobal(action string) error {
	if err := c.SubscribeSelf(action); err != nil {
		return err
	}
	return c.Publish(Message{
		Action: "proto/sub",
		Payload: map[string]interface{}{
			"action": action,
			"device": "",
		},
	})
}

func (c *Client) SubscribeSelf(action string) error {
	return c.Publish(Message{
		Action: "proto/sub",
		Payload: map[string]interface{}{
			"action": action,
			"device": c.DeviceId,
		},
	})
}
