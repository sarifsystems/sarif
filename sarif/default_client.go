// Copyright (C) 2019 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sarif

import (
	"errors"
	"strings"
	"sync"
	"time"
)

type subscription struct {
	Action  string        `json:"action,omitempty"`
	Device  string        `json:"device,omitempty"`
	Handler func(Message) `json:"-"`
}

func (s subscription) Matches(msg Message) bool {
	if msg.Destination != s.Device {
		return false
	}
	if !msg.IsAction(s.Action) {
		return false
	}
	return true
}

type defaultClient struct {
	deviceId         string
	Info             ClientInfo
	RequestTimeout   time.Duration
	HandleConcurrent bool

	conn    Connection
	handler func(Message)
	subs    []subscription

	reqMutex *sync.Mutex
	requests map[string]chan Message
}

func NewClient(ci ClientInfo) Client {
	c := &defaultClient{
		deviceId:         ci.Name,
		RequestTimeout:   30 * time.Second,
		HandleConcurrent: true,

		subs: make([]subscription, 0),

		reqMutex: &sync.Mutex{},
		requests: make(map[string]chan Message),
	}
	return c
}

func (c *defaultClient) DeviceId() string {
	return c.deviceId
}

func (c *defaultClient) Connect(conn Connection) error {
	c.conn = conn

	if c.Info.Auth != "" {
		c.Publish(CreateMessage("proto/hi", c.Info))
	}
	for _, sub := range c.subs {
		c.Subscribe(sub.Action, sub.Device, sub.Handler)
	}
	if err := c.Subscribe("", c.deviceId, nil); err != nil {
		return err
	}
	if err := c.Subscribe("ping", "", c.handlePing); err != nil {
		return err
	}

	go c.listen()
	return nil
}

func (c *defaultClient) Disconnect() error {
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

func (c *defaultClient) listen() error {
	msgs, err := c.conn.Consume()
	if err != nil {
		return err
	}

	for msg := range msgs {
		if c.HandleConcurrent {
			go c.handle(msg)
		} else {
			c.handle(msg)
		}
	}

	return errors.New("Consume stopped")
}

func (c *defaultClient) fillMessage(msg *Message) {
	if msg.Version == "" {
		msg.Version = VERSION
	}
	if msg.Id == "" {
		msg.Id = GenerateId()
	}
	if msg.Source == "" {
		msg.Source = c.deviceId
	}
}

func (c *defaultClient) Publish(msg Message) error {
	c.fillMessage(&msg)
	return c.conn.Publish(msg)
}

func (c *defaultClient) handle(msg Message) {
	if ok := c.resolveRequest(msg.CorrId, msg); ok {
		return
	}

	for _, s := range c.subs {
		if s.Matches(msg) && s.Handler != nil {
			s.Handler(msg)
		}
	}
}

func (c *defaultClient) handlePing(msg Message) {
	c.Reply(msg, CreateMessage("ack", nil))
}

func (c *defaultClient) internalSubscribe(action, device string, h func(Message)) {
	if device == "" && action != "" {
		c.internalSubscribe(action, c.deviceId, h)
	}
	c.subs = append(c.subs, subscription{
		action,
		device,
		h,
	})
}

func (c *defaultClient) Subscribe(action, device string, h func(Message)) error {
	if device == "self" {
		device = c.DeviceId()
	}
	c.internalSubscribe(action, device, h)
	if err := c.conn.Subscribe(c.DeviceId(), action, device); err != nil {
		return err
	}
	if !strings.HasPrefix(action, "proto/discover/") {
		return c.Subscribe("proto/discover/"+action, "", c.handlePing)
	}
	return nil
}

func (c *defaultClient) Reply(orig, reply Message) error {
	return c.Publish(orig.Reply(reply))
}

func (c *defaultClient) ReplyBadRequest(orig Message, err error) error {
	c.Log("err/badrequest", "Bad Request: "+err.Error(), orig)
	return c.Reply(orig, BadRequest(err))
}

func (c *defaultClient) ReplyInternalError(orig Message, err error) error {
	c.Log("err/internal", "Internal Error: "+err.Error(), orig)
	return c.Reply(orig, InternalError(err))
}

func (c *defaultClient) Request(msg Message) <-chan Message {
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

func (c *defaultClient) SetRequestTimeout(timeout time.Duration) {
	c.RequestTimeout = timeout
}

func (c *defaultClient) resolveRequest(id string, msg Message) bool {
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

func (c *defaultClient) Discover(action string) <-chan Message {
	return c.Request(CreateMessage("proto/discover/"+action, nil))
}

func (c *defaultClient) Log(typ, text string, args ...interface{}) error {
	var pl interface{}
	if len(args) > 0 {
		pl = args[0]
	}
	msg := CreateMessage("log/"+typ, pl)
	msg.Text = text
	return c.Publish(msg)
}
