// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sfproto

import (
	"testing"
	"time"

	"github.com/sarifsystems/sarif/sarif"
)

type single struct {
	action string
	device string
	should bool
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

	tc, other := NewPipe()
	client := sarif.NewClient(sarif.ClientInfo{
		Name: "test",
	})
	client.Connect(wrap(other))

	fired := false
	client.Subscribe("ping", "one", func(msg sarif.Message) {
		fired = true
	})

	for _, test := range tests {
		fired = false
		tc.Write(sarif.Message{
			Action:      test.action,
			Destination: test.device,
		})
		time.Sleep(10 * time.Millisecond)
		if test.should && !fired {
			t.Error("did not fire", test)
		}
		if !test.should && fired {
			t.Error("should not fire", test)
		}
	}
}

func TestClientRequest(t *testing.T) {
	aconn, bconn := NewPipe()
	a := sarif.NewClient(sarif.ClientInfo{
		Name: "a",
	})
	b := sarif.NewClient(sarif.ClientInfo{
		Name: "b",
	})
	a.Connect(wrap(bconn))
	b.Connect(wrap(aconn))
	b.SetRequestTimeout(100 * time.Millisecond)

	a.Subscribe("hello_a", "", func(msg sarif.Message) {
		a.Reply(msg, sarif.Message{
			Action: "hi",
		})
	})

	msg, ok := <-b.Request(sarif.Message{
		Action: "hello_a",
	})
	if !ok {
		t.Fatal("A did not respond")
	}
	if msg.Action != "hi" {
		t.Log(msg)
		t.Fatal("did not receive correct response")
	}

	msg, ok = <-b.Request(sarif.Message{
		Action: "hello_no_one",
	})
	if ok {
		t.Fatal("expected no response, got", msg)
	}
}
