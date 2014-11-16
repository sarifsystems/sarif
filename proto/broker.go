// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import "strings"

// Broker dispatches messages to connections based on their subscriptions.
type Broker struct {
	conns    map[*brokerConn]struct{}
	dups     []string
	dupIndex int
	Log      Logger
	trace    bool
}

// NewBroker returns a new broker that dispatches messages.
func NewBroker() *Broker {
	return &Broker{
		make(map[*brokerConn]struct{}, 0),
		make([]string, 128),
		0,
		defaultLog,
		false,
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

// ListenOnConn starts listening on the connection for incoming messages and
// sends outgoing messages based on its subscriptions. The call blocks until
// an error is received, for example when the connection is closed.
func (b *Broker) ListenOnConn(conn Conn) error {
	c := &brokerConn{
		b,
		make(map[string]struct{}, 0),
		make(chan error),
		conn,
	}
	b.conns[c] = struct{}{}

	go func() {
		for {
			msg, err := conn.Read()
			if err != nil {
				c.errs <- err
				return
			}
			c.publish(msg)
		}
	}()
	return <-c.errs
}

// ListenOnBridge forms a bridge between two brokers by transmitting all
// messages in both directions, regardless of subscriptions. The call blocks
// until an error is received.
func (b *Broker) ListenOnBridge(conn Conn) error {
	if err := conn.Write(Subscribe("", "")); err != nil {
		return err
	}

	c := &brokerConn{
		b,
		make(map[string]struct{}, 0),
		make(chan error),
		conn,
	}
	b.conns[c] = struct{}{}
	c.publish(Subscribe("", ""))

	go func() {
		for {
			msg, err := conn.Read()
			if err != nil {
				c.errs <- err
				return
			}
			c.publish(msg)
		}
	}()
	return <-c.errs
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
			return err
		}
		go func() {
			b.Log.Infoln("[broker] connection accepted:", l.Addr())
			err := b.ListenOnConn(conn)
			b.Log.Errorln("[broker] connection closed: ", err)
		}()
	}
}

// Publish publishes a message to all client connections that are subscribed
// to it.
func (b *Broker) Publish(msg Message) {
	if b.checkDuplicate(msg.Id) {
		b.Log.Debugln("[broker] ignore duplicate:", msg)
		return
	}

	if b.trace {
		raw, _ := msg.Encode()
		b.Log.Debugln("[broker] publish:", string(raw))
	}

	topic := getTopic(msg.Action, msg.Destination)
	topic = "/" + strings.TrimLeft(topic, "/")
	t := ""
	for _, part := range strings.Split(topic, "/") {
		t += part
		for c := range b.conns {
			if _, ok := c.subs[t]; ok {
				go func(c *brokerConn) {
					if err := c.conn.Write(msg); err != nil {
						c.errs <- err
					}
				}(c)
			}
		}
		if t != "" {
			t += "/"
		}
	}
}

type brokerConn struct {
	broker *Broker
	subs   map[string]struct{}
	errs   chan error
	conn   Conn
}

func (c *brokerConn) subscribe(topic string) {
	unsubs := make([]string, 0)
	for sub := range c.subs {
		// Already subscribed? Nothing to do.
		if strings.HasPrefix(topic+"/", sub+"/") {
			return
		}
		// Already subscribed to sub topic? Unsubscribe from it.
		if strings.HasPrefix(sub+"/", topic+"/") {
			unsubs = append(unsubs, sub)
		}
	}

	for _, sub := range unsubs {
		c.unsubscribe(sub)
	}

	c.subs[topic] = struct{}{}
}

func (c *brokerConn) unsubscribe(topic string) {
	delete(c.subs, topic)
}

func (c *brokerConn) publish(msg Message) {
	if msg.IsAction("proto/sub") {
		var sub subscription
		if err := msg.DecodePayload(&sub); err == nil {
			topic := getTopic(sub.Action, sub.Device)
			c.subscribe(topic)
		}
	}

	c.broker.Publish(msg)
}
