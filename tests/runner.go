// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package tests

import (
	"sync"
	"testing"
	"time"

	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/transports/sfproto"
)

func ShouldBeAction(actual interface{}, expected ...interface{}) string {
	action, ok := expected[0].(string)
	if !ok {
		return "Has no expected action."
	}

	msg, ok := actual.(sarif.Message)
	if !ok {
		return "Is not a valid Sarif message."
	}
	if msg.Id == "" && msg.Action == "" {
		return "Does not contain any message."
	}

	if msg.IsAction(action) {
		return ""
	}
	return "Has unexpected action '" + msg.Action + "'"
}

type TestRunner struct {
	*testing.T
	conn        sfproto.Conn
	WaitTimeout time.Duration
	IgnoreSubs  bool
	Id          string

	Received chan sarif.Message
	RecMutex sync.Mutex
}

func NewTestRunner(t *testing.T) *TestRunner {
	return &TestRunner{
		T:           t,
		WaitTimeout: time.Second,
		IgnoreSubs:  true,
		Id:          "testutils-" + sarif.GenerateId(),

		Received: make(chan sarif.Message, 5),
		RecMutex: sync.Mutex{},
	}
}

func (t *TestRunner) UseConn(conn sfproto.Conn) {
	t.conn = conn
	t.Publish(sarif.CreateMessage("proto/sub", map[string]string{
		"device": t.Id,
	}))
	go t.listen()
}

func (t *TestRunner) Subscribe(action string) {
	t.Publish(sarif.CreateMessage("proto/sub", map[string]string{
		"action": action,
	}))
}

func (t *TestRunner) listen() {
	for {
		msg, err := t.conn.Read()
		if err != nil {
			t.T.Fatal(err)
		}
		if t.IgnoreSubs && msg.IsAction("proto/sub") {
			continue
		}
		t.RecMutex.Lock()
		t.Received <- msg
		t.RecMutex.Unlock()
	}
}

func (t *TestRunner) Publish(msg sarif.Message) {
	if msg.Source == "" {
		msg.Source = t.Id
	}
	if err := t.conn.Write(msg); err != nil {
		t.T.Fatal(err)
	}
}

func (t *TestRunner) Wait() {
	time.Sleep(100 * time.Millisecond)
}

func (t *TestRunner) When(msgs ...sarif.Message) {
	for _, msg := range msgs {
		t.Publish(msg)
	}
}

func (t *TestRunner) Expect() sarif.Message {
	select {
	case msg := <-t.Received:
		return msg
	case <-time.After(t.WaitTimeout):
		return sarif.Message{}
	}
}
