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

	conn                    Conn
	handler                 func(Message)
	log                     Logger
	subs                    []subscription
	onConnectionLostHandler func(error)

	reqMutex *sync.Mutex
	requests map[string]chan Message
}

func NewClient(deviceId string) *Client {
	c := &Client{
		deviceId,
		2 * time.Second,

		nil,
		nil,
		defaultLog,
		make([]subscription, 0),
		nil,

		&sync.Mutex{},
		make(map[string]chan Message),
	}
	c.internalSubscribe("", c.DeviceId, nil)
	c.internalSubscribe("ping", "", c.handlePing)
	return c
}

func (c *Client) OnConnectionLost(f func(error)) {
	c.onConnectionLostHandler = f
}

func (c *Client) Dial(cfg *NetConfig) error {
	conn, err := Dial(cfg)
	if err != nil {
		return err
	}
	return c.Connect(conn)
}

func (c *Client) Connect(conn Conn) error {
	c.conn = conn
	if err := c.Publish(CreateMessage("proto/subs", c.subs)); err != nil {
		return err
	}

	go c.listen(conn)
	return nil
}

func (c *Client) Disconnect() error {
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

func (c *Client) listen(conn Conn) error {
	for {
		msg, err := c.conn.Read()
		if err != nil {
			c.conn = nil
			c.log.Errorln("[client] read:", err)
			if c.onConnectionLostHandler != nil {
				c.onConnectionLostHandler(err)
			}
			return err
		}
		go c.handle(msg)
	}
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
	if c.conn == nil {
		err := errors.New("not connected")
		c.log.Errorf("[client %s] publish error: ", c.DeviceId, err)
		return err
	}
	if err := c.conn.Write(msg); err != nil {
		c.log.Errorf("[client %s] publish error: %v", c.DeviceId, err)
		return err
	}
	return nil
}

func (c *Client) handle(msg Message) {
	if ok := c.resolveRequest(msg.CorrId, msg); ok {
		return
	}

	for _, s := range c.subs {
		if s.Matches(msg) && s.Handler != nil {
			s.Handler(msg)
		}
	}
}

func (c *Client) handlePing(msg Message) {
	c.log.Debugf("%s got ping", c.DeviceId)
	c.Reply(msg, CreateMessage("ack", nil))
}

func (c *Client) internalSubscribe(action, device string, h func(Message)) {
	if device == "" && action != "" {
		c.internalSubscribe(action, c.DeviceId, h)
	}
	c.subs = append(c.subs, subscription{
		action,
		device,
		h,
	})
}

func (c *Client) Subscribe(action, device string, h func(Message)) error {
	if h == nil {
		return errors.New("Invalid argument: no handler specified")
	}
	if device == "self" {
		device = c.DeviceId
	}
	c.internalSubscribe(action, device, h)
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

	if msg.Action != "" {
		ch <- msg
	} else {
		delete(c.requests, id)
		close(ch)
	}
	return true
}
