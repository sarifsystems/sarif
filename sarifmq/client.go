// Copyright (C) 2019 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sarifmq

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/sfproto"
	"github.com/streadway/amqp"
)

type subscription struct {
	Action  string              `json:"action,omitempty"`
	Device  string              `json:"device,omitempty"`
	Handler func(sarif.Message) `json:"-"`
}

func (s subscription) Matches(msg sarif.Message) bool {
	if msg.Destination != s.Device {
		return false
	}
	if !msg.IsAction(s.Action) {
		return false
	}
	return true
}

type Client struct {
	deviceId         string
	Info             sarif.ClientInfo
	RequestTimeout   time.Duration
	HandleConcurrent bool

	conn                    *amqp.Connection
	channel                 *amqp.Channel
	queue                   amqp.Queue
	handler                 func(sarif.Message)
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

		subs: make([]subscription, 0),

		reqMutex: &sync.Mutex{},
		requests: make(map[string]chan sarif.Message),
	}
	c.internalSubscribe("", c.deviceId, nil)
	c.internalSubscribe("ping", "", c.handlePing)
	return c
}

func (c *Client) DeviceId() string {
	return c.deviceId
}

func (c *Client) OnConnectionLost(f func(error)) {
	c.onConnectionLostHandler = f
}

func (c *Client) Dial(cfg *sfproto.NetConfig) error {
	conn, err := amqp.Dial(cfg.Address)
	if err != nil {
		return err
	}
	return c.Connect(conn)
}

func (c *Client) Connect(conn *amqp.Connection) error {
	c.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	c.channel = ch

	err = c.channel.ExchangeDeclare(
		"sarif",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	q, err := c.channel.QueueDeclare(
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
	c.queue = q

	if c.Info.Auth != "" {
		c.Publish(sarif.CreateMessage("proto/hi", c.Info))
	}

	go c.listen()
	return nil
}

func (c *Client) Disconnect() error {
	if c.conn == nil {
		return nil
	}
	if err := c.channel.Close(); err != nil {
		return err
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

func (c *Client) listen() error {
	msgs, err := c.channel.Consume(
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

	for amsg := range msgs {
		var msg sarif.Message
		err := json.Unmarshal(amsg.Body, &msg)
		if err != nil {
			c.conn = nil
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

	return errors.New("Consume stopped")
}

func (c *Client) fillMessage(msg *sarif.Message) {
	if msg.Version == "" {
		msg.Version = sarif.VERSION
	}
	if msg.Id == "" {
		msg.Id = sarif.GenerateId()
	}
	if msg.Source == "" {
		msg.Source = c.deviceId
	}
}

func (c *Client) Publish(msg sarif.Message) error {
	c.fillMessage(&msg)
	if c.conn == nil {
		err := errors.New("not connected")
		return err
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	topic := getTopic(msg.Action, msg.Destination)

	return c.channel.Publish(
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
		c.internalSubscribe(action, c.deviceId, h)
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
		device = c.deviceId
	}
	c.internalSubscribe(action, device, h)
	topic := getTopic(action, device) + ".#"
	fmt.Println(topic)
	if err := c.channel.QueueBind(c.queue.Name, topic, "sarif", false, nil); err != nil {
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
