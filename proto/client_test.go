// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"testing"
	"time"
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
	client := NewClient("test", other)

	fired := false
	client.Subscribe("ping", "one", func(msg Message) {
		fired = true
	})

	for _, test := range tests {
		fired = false
		tc.Write(Message{
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
