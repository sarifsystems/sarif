// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package testutils provides a BDD test framework for stark services.
package testutils

import (
	"sync"
	"testing"
	"time"

	"github.com/xconstruct/stark/proto"
)

type Tester struct {
	*testing.T
	conn        proto.Conn
	WaitTimeout time.Duration
	IgnoreSubs  bool
	Id          string

	Unit       string
	Behaviour  string
	Received   chan proto.Message
	ExpectCurr *proto.Message
	RecMutex   sync.Mutex
}

func New(t *testing.T) *Tester {
	return &Tester{
		T:           t,
		WaitTimeout: time.Second,
		IgnoreSubs:  true,
		Id:          "testutils-" + proto.GenerateId(),

		Received: make(chan proto.Message, 5),
		RecMutex: sync.Mutex{},
	}
}

func (t *Tester) UseConn(conn proto.Conn) {
	t.conn = conn
	go t.listen()
}

func (t *Tester) CreateConn() proto.Conn {
	a, b := proto.NewPipe()
	t.UseConn(a)
	return b
}

func (t *Tester) listen() {
	for {
		msg, err := t.conn.Read()
		if err != nil {
			t.T.Fatal(err)
		}
		if t.IgnoreSubs && msg.IsAction("proto/sub") {
			continue
		}
		if msg.Source == t.Id {
			continue
		}
		t.RecMutex.Lock()
		t.Received <- msg
		t.RecMutex.Unlock()
	}
}

func (t *Tester) Publish(msg proto.Message) {
	if msg.Source == "" {
		msg.Source = t.Id
	}
	if err := t.conn.Write(msg); err != nil {
		t.T.Fatal(err)
	}
}

func (t *Tester) Wait() {
	time.Sleep(50 * time.Millisecond)
}

func (t *Tester) Reset() {
	t.RecMutex.Lock()
	t.Received = make(chan proto.Message, 20)
	t.RecMutex.Unlock()
}

func (t *Tester) Describe(unit string, f func()) {
	t.Unit = unit
	f()
}

func (t *Tester) It(behaviour string, f func()) {
	t.Reset()
	t.Behaviour = behaviour
	f()

	if t.HasReplies() {
		t.T.Log(t.Received)
		t.T.Fatal(t.Unit, t.Behaviour+": still replies left")
	}
}

func (t *Tester) When(msgs ...proto.Message) {
	for _, msg := range msgs {
		t.Publish(msg)
	}
}

func (t *Tester) HasReplies() bool {
	return len(t.Received) > 0
}

func (t *Tester) Expect(f func(proto.Message)) {
	if t.ExpectCurr != nil {
		f(*t.ExpectCurr)
		return
	}

	select {
	case msg := <-t.Received:
		t.ExpectCurr = &msg
		f(msg)
		t.ExpectCurr = nil
	case <-time.After(t.WaitTimeout):
		t.T.Fatal(t.Unit, t.Behaviour+": no message received")
	}
}

func (t *Tester) ExpectAction(action string) {
	t.Expect(func(msg proto.Message) {
		if !msg.IsAction(action) {
			t.T.Fatal(t.Unit, t.Behaviour+": expected action", action, "not", msg.Action+":", msg.Text)
		}
	})
}

func (t *Tester) ExpectText(text string) {
	t.Expect(func(msg proto.Message) {
		if msg.Text != text {
			t.T.Fatalf("%s %s: expected text '%s', not '%s'", t.Unit, t.Behaviour, text, msg.Text)
		}
	})
}
