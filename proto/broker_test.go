// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"testing"
	"time"
)

type brokerMultiEp struct {
	action    string
	device    string
	oneShould bool
	twoShould bool
}

func TestBrokerMultiple(t *testing.T) {
	tests := []brokerMultiEp{
		{"ping", "one", true, false},
		{"ping", "two", false, true},
		{"ping", "", false, true},
		{"ack", "one", false, false},
		{"ack", "two", false, false},
		{"ack", "", false, false},
	}

	oneFired, twoFired := false, false
	b := NewBroker()

	// Setup first client
	epOne, oneOther := NewPipe()
	go func() {
		for {
			_, err := epOne.Read()
			if err != nil {
				t.Fatal(err)
			}
			oneFired = true
		}
	}()
	go func() {
		t.Fatal(b.ListenOnConn(oneOther))
	}()
	err := epOne.Write(CreateMessage("proto/sub", subscription{"ping", "one", nil}))
	if err != nil {
		t.Fatal(err)
	}

	// Setup second client
	epTwo, twoOther := NewPipe()
	go func() {
		for {
			_, err := epTwo.Read()
			if err != nil {
				t.Fatal(err)
			}
			twoFired = true
		}
	}()
	go func() {
		t.Fatal(b.ListenOnConn(twoOther))
	}()
	err = epTwo.Write(CreateMessage("proto/sub", subscription{"ping", "two", nil}))
	if err != nil {
		t.Fatal(err)
	}
	err = epTwo.Write(CreateMessage("proto/sub", subscription{"ping", "", nil}))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(10 * time.Millisecond)

	for i, test := range tests {
		oneFired, twoFired = false, false
		err := epOne.Write(Message{
			Id:          GenerateId(),
			Action:      test.action,
			Destination: test.device,
		})
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
		if test.oneShould && !oneFired {
			t.Error(i, "one did not fire", test)
		}
		if !test.oneShould && oneFired {
			t.Error(i, "one should not fire", test)
		}
		if test.twoShould && !twoFired {
			t.Error(i, "two did not fire", test)
		}
		if !test.twoShould && twoFired {
			t.Error(i, "two should not fire", test)
		}
	}
}
