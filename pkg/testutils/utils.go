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
	Conn      proto.Conn
	localConn proto.Conn

	Unit string
	Test Test
}

type Test struct {
	Behaviour string
	Received  []proto.Message
	Waited    bool
}

func New(t *testing.T) *Tester {
	a, b := proto.NewPipe()
	st := &Tester{
		t,
		a,
		b,

		"",
		Test{},
	}
	go st.listen()
	return st
}

func (t *Tester) listen() {
	for {
		msg, err := t.localConn.Read()
		if err != nil {
			t.T.Fatal(err)
		}
		t.Test.Received = append(t.Test.Received, msg)
	}
}

func (t *Tester) Publish(msg proto.Message) {
	if err := t.localConn.Write(msg); err != nil {
		t.T.Fatal(err)
	}
}

func (t *Tester) Wait() {
	time.Sleep(50 * time.Millisecond)
}

func (t *Tester) Reset() {
	t.Test = Test{}
}

func (t *Tester) Describe(unit string, f func()) {
	t.Unit = unit
	f()
}

func (t *Tester) It(behaviour string, f func()) {
	t.Reset()
	t.Test.Behaviour = behaviour
	f()

	if t.HasReplies() {
		t.T.Log(t.Test.Received)
		t.T.Fatal(t.Unit, t.Test.Behaviour+": still replies left")
	}
}

func (t *Tester) When(msgs ...proto.Message) {
	for _, msg := range msgs {
		t.Publish(msg)
	}
}

func (t *Tester) HasReplies() bool {
	return t.Test.Received != nil && len(t.Test.Received) > 0
}

func (t *Tester) Expect(f func(proto.Message)) {
	if !t.Test.Waited {
		t.Test.Waited = true
		t.Wait()
	}

	if !t.HasReplies() {
		t.T.Fatal(t.Unit, t.Test.Behaviour+": no message received")
	}
	f(t.Test.Received[0])
	t.Test.Received = t.Test.Received[1:]
}

func (t *Tester) DiscardTheRest() {
	t.Test.Received = nil
}
