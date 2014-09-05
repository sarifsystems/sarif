// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"errors"

	"github.com/xconstruct/stark/log"
)

type Client struct {
	DeviceId string
	endpoint Endpoint
	handler  Handler
	log      log.Interface
	subs     []Subscription
}

func NewClient(deviceId string, e Endpoint) *Client {
	c := &Client{
		deviceId,
		e,
		nil,
		log.Default,
		make([]Subscription, 0),
	}
	c.endpoint = e
	e.RegisterHandler(c.handle)
	c.Subscribe("ping", "", c.handlePing)
	return c
}

func (c *Client) SetLogger(l log.Interface) {
	c.log = l
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
	for _, s := range c.subs {
		if s.Matches(msg) {
			s.Handler(msg)
		}
	}
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

func (c *Client) Subscribe(action, device string, h Handler) error {
	if h == nil {
		return errors.New("Invalid argument: no handler specified")
	}
	if device == "" && action != "" {
		if err := c.Subscribe(action, c.DeviceId, h); err != nil {
			return err
		}
	}
	if device == "self" {
		device = c.DeviceId
	}

	c.subs = append(c.subs, Subscription{
		action,
		device,
		h,
	})
	return c.Publish(Message{
		Action: "proto/sub",
		Payload: map[string]interface{}{
			"action": action,
			"device": device,
		},
	})
}
