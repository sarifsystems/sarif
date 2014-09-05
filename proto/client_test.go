// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"testing"
)

type single struct {
	action string
	device string
	should bool
}

type testConn struct {
	last Message
	send Handler
}

func (c *testConn) Publish(msg Message) error {
	c.last = msg
	return nil
}

func (c *testConn) RegisterHandler(h Handler) {
	c.send = h
}

func TestClientSingle(t *testing.T) {
	tests := []single{
		{"ping", "one", true},
		{"ping", "two", false},
		{"ping", "", false},
		{"ack", "one", false},
		{"ack", "two", false},
		{"ack", "", false},
	}

	tc := &testConn{}
	client := NewClient("test", tc)

	fired := false
	client.Subscribe("ping", "one", func(msg Message) {
		fired = true
	})
	if tc.send == nil {
		t.Fatal("client should have registed with the testconn")
	}

	for _, test := range tests {
		fired = false
		tc.send(Message{
			Action:      test.action,
			Destination: test.device,
		})
		if test.should && !fired {
			t.Error("did not fire", test)
		}
		if !test.should && fired {
			t.Error("should not fire", test)
		}
	}
}
