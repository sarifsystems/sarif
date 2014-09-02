// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mux

import (
	"testing"

	"github.com/xconstruct/stark/proto"
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

	mux := New()
	fired := false
	mux.RegisterHandler("ping", "one", func(msg proto.Message) {
		fired = true
	})
	for _, test := range tests {
		fired = false
		mux.Handle(proto.Message{
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

	mux := New()
	oneFired, twoFired := false, false
	mux.RegisterHandler("ping", "one", func(msg proto.Message) {
		oneFired = true
	})
	mux.RegisterHandler("ping", "two", func(msg proto.Message) {
		twoFired = true
	})
	mux.RegisterHandler("ping", "", func(msg proto.Message) {
		twoFired = true
	})
	for _, test := range tests {
		oneFired, twoFired = false, false
		mux.Handle(proto.Message{
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
