// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import "io"

// Broker dispatches messages to connections based on their subscriptions.
type Broker struct {
	subs     *subTree
	dups     []string
	dupIndex int
	Log      Logger
	trace    bool
}

// NewBroker returns a new broker that dispatches messages.
func NewBroker() *Broker {
	return &Broker{
		newSubtree(),
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
	topics := b.subs.GetTopics("", []string{})
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
func (b *Broker) publish(msg Message) {
	if b.checkDuplicate(msg.Id) {
		if b.trace {
			b.Log.Debugln("[broker] ignore duplicate:", msg)
		}
		return
	}

	if b.trace {
		raw, _ := msg.Encode()
		b.Log.Debugln("[broker] publish:", string(raw))
	}

	topic := getTopic(msg.Action, msg.Destination)
	b.subs.Call(topicParts(topic), func(c Writer) {
		go c.Write(msg)
	})
}

func (b *Broker) PrintSubtree(w io.Writer) error {
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
	c.broker.subs.Unsubscribe(nil, c)
	return c.Conn.Close()
}

func (c *brokerConn) ListenLoop() {
	for {
		msg, err := c.Conn.Read()
		if err != nil {
			return
		}
		c.Publish(msg)
	}
}

func (c *brokerConn) Subscribe(topic string) {
	c.broker.subs.Subscribe(topicParts(topic), c)
}

func (c *brokerConn) Unsubscribe(topic string) {
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
	}

	c.broker.publish(msg)
}

func (c *brokerConn) String() string {
	if v, ok := c.Conn.(stringer); ok {
		return v.String()
	}
	return ""
}
