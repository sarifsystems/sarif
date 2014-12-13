// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"errors"
	"sync"
	"time"
)

type Client struct {
	DeviceId       string
	RequestTimeout time.Duration

	conn    Conn
	handler Handler
	log     Logger
	subs    []subscription

	reqMutex *sync.Mutex
	requests map[string]chan Message
}

func NewClient(deviceId string, e Conn) *Client {
	c := &Client{
		deviceId,
		2 * time.Second,

		e,
		nil,
		defaultLog,
		make([]subscription, 0),
		&sync.Mutex{},
		make(map[string]chan Message),
	}
	c.conn = e
	go func() {
		for {
			msg, err := e.Read()
			if err != nil {
				c.log.Errorln("[client] read:", err)
				return
			}
			go c.handle(msg)
		}

	}()
	c.Subscribe("ping", "", c.handlePing)
	return c
}

func (c *Client) SetLogger(l Logger) {
	c.log = l
}

func (c *Client) fillMessage(msg *Message) {
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
	c.fillMessage(&msg)
	if err := c.conn.Write(msg); err != nil {
		c.log.Errorf("[client %s] publish error: %v, %v", c.DeviceId, err)
		return err
	}
	return nil
}

func (c *Client) handle(msg Message) {
	if ok := c.resolveRequest(msg.CorrId, msg); ok {
		return
	}

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
		c.log.Warnf("%s", err)
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

	c.subs = append(c.subs, subscription{
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
	c.log.Errorf("[client %s] internal error: %v, %v", c.DeviceId, orig, err)
	reply := orig.Reply(InternalError(err))
	return c.Publish(reply)
}

func (c *Client) Request(msg Message) <-chan Message {
	c.fillMessage(&msg)
	ch := make(chan Message, 1)
	if err := c.Publish(msg); err != nil {
		close(ch)
		return ch
	}

	go func(id string) {
		time.Sleep(c.RequestTimeout)
		c.resolveRequest(id, Message{})
	}(msg.Id)

	c.reqMutex.Lock()
	defer c.reqMutex.Unlock()
	c.requests[msg.Id] = ch
	return ch
}

func (c *Client) resolveRequest(id string, msg Message) bool {
	if id == "" {
		return false
	}

	c.reqMutex.Lock()
	defer c.reqMutex.Unlock()
	ch, ok := c.requests[id]
	if !ok {
		return false
	}

	if msg.Id != "" {
		ch <- msg
	} else {
		delete(c.requests, id)
		close(ch)
	}
	return true
}
