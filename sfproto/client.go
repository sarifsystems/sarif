// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sfproto

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/sarifsystems/sarif/sarif"
)

type Client struct {
	deviceId         string
	Info             sarif.ClientInfo
	RequestTimeout   time.Duration
	HandleConcurrent bool

	conn                    Conn
	handler                 func(sarif.Message)
	log                     Logger
	subs                    []subscription
	onConnectionLostHandler func(error)

	reqMutex *sync.Mutex
	requests map[string]chan sarif.Message
}

func NewClient(deviceId string) *Client {
	c := &Client{
		deviceId:         deviceId,
		RequestTimeout:   30 * time.Second,
		HandleConcurrent: true,

		log:  defaultLog,
		subs: make([]subscription, 0),

		reqMutex: &sync.Mutex{},
		requests: make(map[string]chan sarif.Message),
	}
	c.internalSubscribe("", c.DeviceId(), nil)
	c.internalSubscribe("ping", "", c.handlePing)
	return c
}

func (c *Client) DeviceId() string {
	return c.deviceId
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
	if c.Info.Auth != "" {
		c.Publish(sarif.CreateMessage("proto/hi", c.Info))
	}
	if err := c.Publish(sarif.CreateMessage("proto/subs", c.subs)); err != nil {
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
		msg, err := conn.Read()
		if err != nil {
			c.conn = nil
			c.log.Errorln("[client] read:", err)
			if c.onConnectionLostHandler != nil {
				c.onConnectionLostHandler(err)
			}
			return err
		}
		if c.HandleConcurrent {
			go c.handle(msg)
		} else {
			c.handle(msg)
		}
	}
}

func (c *Client) SetLogger(l Logger) {
	c.log = l
}

func (c *Client) fillMessage(msg *sarif.Message) {
	if msg.Version == "" {
		msg.Version = sarif.VERSION
	}
	if msg.Id == "" {
		msg.Id = sarif.GenerateId()
	}
	if msg.Source == "" {
		msg.Source = c.DeviceId()
	}
}

func (c *Client) Publish(msg sarif.Message) error {
	c.fillMessage(&msg)
	if c.conn == nil {
		err := errors.New("not connected")
		c.log.Errorf("[client %s] publish error: %v", c.DeviceId, err)
		return err
	}
	if err := c.conn.Write(msg); err != nil {
		c.log.Errorf("[client %s] publish error: %v", c.DeviceId, err)
		return err
	}
	return nil
}

func (c *Client) handle(msg sarif.Message) {
	if ok := c.resolveRequest(msg.CorrId, msg); ok {
		return
	}

	for _, s := range c.subs {
		if s.Matches(msg) && s.Handler != nil {
			s.Handler(msg)
		}
	}
}

func (c *Client) handlePing(msg sarif.Message) {
	c.Reply(msg, sarif.CreateMessage("ack", nil))
}

func (c *Client) internalSubscribe(action, device string, h func(sarif.Message)) {
	if device == "" && action != "" {
		c.internalSubscribe(action, c.DeviceId(), h)
	}
	c.subs = append(c.subs, subscription{
		action,
		device,
		h,
	})
}

func (c *Client) Subscribe(action, device string, h func(sarif.Message)) error {
	if h == nil {
		return errors.New("Invalid argument: no handler specified")
	}
	if device == "self" {
		device = c.DeviceId()
	}
	c.internalSubscribe(action, device, h)
	if err := c.Publish(Subscribe(action, device)); err != nil {
		return err
	}
	if !strings.HasPrefix(action, "proto/discover/") {
		return c.Subscribe("proto/discover/"+action, "", c.handlePing)
	}
	return nil
}

func (c *Client) Reply(orig, reply sarif.Message) error {
	return c.Publish(orig.Reply(reply))
}

func (c *Client) ReplyBadRequest(orig sarif.Message, err error) error {
	c.Log("err/badrequest", "Bad Request: "+err.Error(), orig)
	return c.Reply(orig, sarif.BadRequest(err))
}

func (c *Client) ReplyInternalError(orig sarif.Message, err error) error {
	c.Log("err/internal", "Internal Error: "+err.Error(), orig)
	return c.Reply(orig, sarif.InternalError(err))
}

func (c *Client) Request(msg sarif.Message) <-chan sarif.Message {
	c.fillMessage(&msg)
	ch := make(chan sarif.Message, 1)
	if err := c.Publish(msg); err != nil {
		close(ch)
		return ch
	}

	go func(id string) {
		time.Sleep(c.RequestTimeout)
		c.resolveRequest(id, sarif.Message{})
	}(msg.Id)

	c.reqMutex.Lock()
	defer c.reqMutex.Unlock()
	c.requests[msg.Id] = ch
	return ch
}

func (c *Client) resolveRequest(id string, msg sarif.Message) bool {
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

func (c *Client) Discover(action string) <-chan sarif.Message {
	return c.Request(sarif.CreateMessage("proto/discover/"+action, nil))
}

func (c *Client) Log(typ, text string, args ...interface{}) error {
	var pl interface{}
	if len(args) > 0 {
		pl = args[0]
	}
	msg := sarif.CreateMessage("log/"+typ, pl)
	msg.Text = text
	return c.Publish(msg)
}
