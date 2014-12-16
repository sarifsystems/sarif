// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package testutils

import (
	"testing"
	"time"

	"github.com/xconstruct/stark/proto"
)

type Tester struct {
	*testing.T
	conn        proto.Conn
	WaitTimeout time.Duration
	IgnoreSubs  bool

	Unit      string
	Behaviour string
	Received  chan proto.Message
}

func New(t *testing.T) *Tester {
	return &Tester{
		t,
		nil,
		time.Second,
		true,

		"",
		"",
		make(chan proto.Message, 5),
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
		t.Received <- msg
	}
}

func (t *Tester) Publish(msg proto.Message) {
	if err := t.conn.Write(msg); err != nil {
		t.T.Fatal(err)
	}
}

func (t *Tester) Wait() {
	time.Sleep(50 * time.Millisecond)
}

func (t *Tester) Reset() {
	t.Received = make(chan proto.Message, 20)
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
	select {
	case msg := <-t.Received:
		f(msg)
	case <-time.After(t.WaitTimeout):
		t.T.Fatal(t.Unit, t.Behaviour+": no message received")
	}
}

func (t *Tester) ExpectAction(action string) {
	t.Expect(func(msg proto.Message) {
		if !msg.IsAction(action) {
			t.T.Fatal(t.Unit, t.Behaviour+": expected action", action, "not", msg.Action)
		}
	})
}
