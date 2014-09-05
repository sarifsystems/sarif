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

func TestMuxSingle(t *testing.T) {
	tests := []single{
		{"ping", "one", true},
		{"ping", "two", false},
		{"ping", "", false},
		{"ack", "one", false},
		{"ack", "two", false},
		{"ack", "", false},
	}

	mux := NewMux()
	fired := false
	mux.RegisterHandler("ping", "one", func(msg Message) {
		fired = true
	})
	for _, test := range tests {
		fired = false
		mux.Handle(Message{
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

type multi struct {
	action    string
	device    string
	oneShould bool
	twoShould bool
}

func TestMuxMultiple(t *testing.T) {
	tests := []multi{
		{"ping", "one", true, false},
		{"ping", "two", false, true},
		{"ping", "", false, true},
		{"ack", "one", false, false},
		{"ack", "two", false, false},
		{"ack", "", false, false},
	}

	mux := NewMux()
	oneFired, twoFired := false, false
	mux.RegisterHandler("ping", "one", func(msg Message) {
		oneFired = true
	})
	mux.RegisterHandler("ping", "two", func(msg Message) {
		twoFired = true
	})
	mux.RegisterHandler("ping", "", func(msg Message) {
		twoFired = true
	})
	for _, test := range tests {
		oneFired, twoFired = false, false
		mux.Handle(Message{
			Action:      test.action,
			Destination: test.device,
		})
		if test.oneShould && !oneFired {
			t.Error("one did not fire", test)
		}
		if !test.oneShould && oneFired {
			t.Error("one should not fire", test)
		}
		if test.twoShould && !twoFired {
			t.Error("two did not fire", test)
		}
		if !test.twoShould && twoFired {
			t.Error("two should not fire", test)
		}
	}
}
