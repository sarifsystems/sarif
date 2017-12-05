// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sarif

import (
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"
)

type canVerify interface {
	IsVerified() bool
}

// Broker dispatches messages to connections based on their subscriptions.
type Broker struct {
	subs          *subTree
	subsLock      sync.RWMutex
	dups          []string
	dupIndex      int
	Log           Logger
	trace         bool
	halfOpenConns map[string]chan bool
	clients       map[string]*ClientInfo
}

// NewBroker returns a new broker that dispatches messages.
func NewBroker() *Broker {
	return &Broker{
		subs:          newSubtree(),
		subsLock:      sync.RWMutex{},
		dups:          make([]string, 1024),
		dupIndex:      0,
		Log:           defaultLog,
		trace:         false,
		halfOpenConns: make(map[string]chan bool),
		clients:       make(map[string]*ClientInfo),
	}
}

// SetLogger sets the default log output.
func (b *Broker) SetLogger(log Logger) {
	b.Log = log
}

// TraceMessages enables debug output of individual messages
func (b *Broker) TraceMessages(enabled bool) {
	b.trace = enabled
}

// SetDuplicateDepth sets the number of stored messages to check for
// duplicates. A zero value disables duplicate checking.
func (b *Broker) SetDuplicateDepth(depth int) {
	b.dups = make([]string, depth)
	b.dupIndex = 0
}

// CheckDuplicate checks for duplicate message IDs and adds the new one
// to the buffer.
func (b *Broker) checkDuplicate(id string) bool {
	if id == "" || len(b.dups) == 0 {
		return false
	}
	for i := b.dupIndex - 1; i >= 0; i-- {
		if b.dups[i] == id {
			return true
		}
	}
	for i := len(b.dups) - 1; i >= b.dupIndex; i-- {
		if b.dups[i] == id {
			return true
		}
	}

	b.dups[b.dupIndex] = id
	b.dupIndex = (b.dupIndex + 1) % len(b.dups)
	return false
}

func (b *Broker) newConn(c Conn) *brokerConn {
	return &brokerConn{
		c,
		b,
		make(map[string]struct{}, 0),
		make(chan error),
	}
}

// ListenOnConn starts listening on the connection for incoming messages and
// sends outgoing messages based on its subscriptions. The call blocks until
// an error is received, for example when the connection is closed.
func (b *Broker) ListenOnConn(conn Conn) error {
	c := b.newConn(conn)
	go c.ListenLoop()
	err := <-c.errs
	c.Close()
	return err
}

// ListenOnBridge forms a bridge between two brokers by transmitting all
// messages in both directions, regardless of subscriptions. The call blocks
// until an error is received.
func (b *Broker) ListenOnBridge(conn Conn) error {
	sub := Subscribe("", "")
	sub.Source = "broker"
	if err := conn.Write(sub); err != nil {
		conn.Close()
		return err
	}

	c := b.newConn(conn)
	c.Publish(sub)
	go c.ListenLoop()
	err := <-c.errs
	c.Close()
	return err
}

// ListenOnGateway forms a client-server-relationship between two brokers by
// transmitting all local messages to the gateway, but receiving only subscribed
// messages in return.
func (b *Broker) ListenOnGateway(conn Conn) error {
	b.subsLock.RLock()
	topics := b.subs.GetTopics("", []string{})
	b.subsLock.RUnlock()

	if len(topics) > 0 {
		subs := make([]subscription, len(topics))
		for i, t := range topics {
			subs[i].Action, subs[i].Device = fromTopic(t)
		}
		sub := CreateMessage("proto/subs", subs)
		sub.Source = "broker"
		if err := conn.Write(sub); err != nil {
			conn.Close()
			return err
		}
	}

	c := b.newConn(conn)
	c.Subscribe("")
	go c.ListenLoop()
	err := <-c.errs
	c.Close()
	return err
}

// NewLocalConn creates a new local connection and starts listening on it.
func (b *Broker) NewLocalConn() Conn {
	one, two := NewPipe()
	go b.ListenOnConn(one)
	return two
}

// Listen starts accepting connections on the specified network and blocks
// until an error is received.
func (b *Broker) Listen(cfg *NetConfig) error {
	l, err := Listen(cfg)
	if err != nil {
		return err
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			b.Log.Errorln("[broker] connect accept failed:", err, l.Addr())
			continue
		}

		go func() {
			b.Log.Infoln("[broker] connection accepted:", l.Addr())
			err := b.AuthenticateAndListenOnConn(cfg.Auth, conn)
			b.Log.Errorln("[broker] connection closed: ", err)
		}()
	}
}

func (b *Broker) AuthenticateAndListenOnConn(auth AuthType, c Conn) error {
	authed := false
	if auth == AuthNone {
		authed = true
	} else {
		if v, ok := c.(canVerify); ok {
			authed = v.IsVerified()
		}
	}

	if !authed && auth != AuthCertificate {
		msg, err := c.Read()
		if err != nil {
			return err
		}
		if !msg.IsAction("proto/hi") {
			return errors.New("Authentication failed: unexpected message " + msg.Action)
		}

		var ci ClientInfo
		msg.DecodePayload(&ci)
		ci.Name = msg.Source
		ci.LastSeen = time.Now()
		b.clients[ci.Name] = &ci
		msg.EncodePayload(ci)
		msg.Id = GenerateId()

		confirm := make(chan bool, 1)
		b.halfOpenConns[msg.Id] = confirm
		b.publish(msg)

		select {
		case <-confirm:
			authed = true
		case <-time.After(time.Minute):
		}
		delete(b.halfOpenConns, msg.Id)
	}

	if !authed {
		c.Close()
		return errors.New("Authentication failed")
	}

	return b.ListenOnConn(c)
}

// Publish publishes a message to all client connections that are subscribed
// to it.
func (b *Broker) publish(msg Message) {
	if b.checkDuplicate(msg.Id) {
		return
	}

	if b.trace {
		raw, _ := json.Marshal(msg)
		b.Log.Debugln("[broker] publish:", string(raw))
	}

	topic := getTopic(msg.Action, msg.Destination)
	b.subsLock.RLock()
	b.subs.Call(topicParts(topic), false, func(c writer) {
		go c.Write(msg)
	})
	b.subsLock.RUnlock()
}

func (b *Broker) PrintSubtree(w io.Writer) error {
	b.subsLock.RLock()
	defer b.subsLock.RUnlock()
	return b.subs.Print(w, 0)
}

type brokerConn struct {
	Conn
	broker *Broker
	subs   map[string]struct{}
	errs   chan error
}

func (c *brokerConn) Write(msg Message) error {
	if err := c.Conn.Write(msg); err != nil {
		c.errs <- err
		return err
	}
	return nil
}

func (c *brokerConn) Read() (Message, error) {
	msg, err := c.Conn.Read()
	if err != nil {
		c.errs <- err
	}
	return msg, err
}

func (c *brokerConn) Close() error {
	c.broker.subsLock.Lock()
	defer c.broker.subsLock.Unlock()
	c.broker.subs.Unsubscribe(nil, c)
	return c.Conn.Close()
}

func (c *brokerConn) ListenLoop() {
	for {
		msg, err := c.Conn.Read()
		if err != nil {
			c.errs <- err
			return
		}
		c.Publish(msg)
	}
}

func (c *brokerConn) Subscribe(topic string) {
	c.broker.subsLock.Lock()
	defer c.broker.subsLock.Unlock()
	c.broker.subs.Subscribe(topicParts(topic), c)
}

func (c *brokerConn) Unsubscribe(topic string) {
	c.broker.subsLock.Lock()
	defer c.broker.subsLock.Unlock()
	c.broker.subs.Unsubscribe(topicParts(topic), c)
}

func (c *brokerConn) Publish(msg Message) {
	switch {
	case msg.IsAction("proto/sub"):
		var sub subscription
		if err := msg.DecodePayload(&sub); err == nil {
			topic := getTopic(sub.Action, sub.Device)
			c.Subscribe(topic)
		}
	case msg.IsAction("proto/unsub"):
		var sub subscription
		if err := msg.DecodePayload(&sub); err == nil {
			topic := getTopic(sub.Action, sub.Device)
			c.Unsubscribe(topic)
		}
		return // TODO: unsub propagation
	case msg.IsAction("proto/subs"):
		var subs []subscription
		if err := msg.DecodePayload(&subs); err == nil {
			for _, sub := range subs {
				topic := getTopic(sub.Action, sub.Device)
				c.Subscribe(topic)
			}
		}
	case msg.IsAction("proto/unsubs"):
		var subs []subscription
		if err := msg.DecodePayload(&subs); err == nil {
			for _, sub := range subs {
				topic := getTopic(sub.Action, sub.Device)
				c.Unsubscribe(topic)
			}
		}
		return // TODO: unsub propagation
	case msg.IsAction("proto/reg"):
		return
	case msg.IsAction("proto/req"):
		return
	case msg.IsAction("proto/allow"):
		if ch, ok := c.broker.halfOpenConns[msg.CorrId]; ok {
			ch <- true
		}
	case msg.IsAction("log"):
		if msg.IsAction("log/err") {
			c.broker.Log.Errorf("[%s] %s - %s", msg.Source, msg.Text, msg.Payload)
		} else {
			c.broker.Log.Infof("[%s] %s - %s", msg.Source, msg.Text, msg.Payload)
		}
	}

	c.broker.publish(msg)

	if ci, ok := c.broker.clients[msg.Source]; ok {
		ci.LastSeen = time.Now()
	}
}

func (c *brokerConn) String() string {
	if v, ok := c.Conn.(interface {
		String() string
	}); ok {
		return v.String()
	}
	return ""
}
