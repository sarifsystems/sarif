// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"errors"
	"fmt"

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
	fmt.Sprintln(c.endpoint)
	if err := c.endpoint.Publish(msg); err != nil {
		c.log.Errorf("[client %s] publish error: %v, %v", c.DeviceId, err)
		return err
	}
	return nil
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
	return c.Publish(Subscribe(action, device))
}

func (c *Client) Reply(orig, reply Message) error {
	return c.Publish(orig.Reply(reply))
}

func (c *Client) ReplyBadRequest(orig Message, err error) error {
	c.log.Warnf("[client %s] bad request: %v, %v", c.DeviceId, orig, err)
	reply := orig.Reply(BadRequest(err))
	return c.Publish(reply)
}

func (c *Client) ReplyInternalError(orig Message, err error) error {
	c.log.Errorln("[client %s] internal error: %v, %v", c.DeviceId, orig, err)
	reply := orig.Reply(InternalError(err))
	return c.Publish(reply)
}
