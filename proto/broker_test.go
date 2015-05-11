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

	one := newTestService("one", t)
	two := newTestService("two", t)
	b := NewBroker()
	go func() {
		t.Fatal(b.ListenOnConn(one.NewLocalConn()))
	}()
	go func() {
		t.Fatal(b.ListenOnConn(two.NewLocalConn()))
	}()
	time.Sleep(10 * time.Millisecond)

	one.Publish(Subscribe("ping", "one"))
	two.Publish(Subscribe("ping", "two"))
	two.Publish(Subscribe("ping", ""))
	time.Sleep(10 * time.Millisecond)

	for i, test := range tests {
		one.Reset()
		two.Reset()
		one.Publish(Message{
			Id:          GenerateId(),
			Action:      test.action,
			Destination: test.device,
		})
		time.Sleep(10 * time.Millisecond)

		if test.oneShould && !one.Fired() {
			t.Error(i, "one did not fire", test)
		}
		if !test.oneShould && one.Fired() {
			t.Error(i, "one should not fire", test)
		}
		if test.twoShould && !two.Fired() {
			t.Error(i, "two did not fire", test)
		}
		if !test.twoShould && two.Fired() {
			t.Error(i, "two should not fire", test)
		}
	}
}

func TestBrokerBridge(t *testing.T) {
	tests := []brokerMultiEp{
		{"ping", "one", true, false},
		{"ping", "two", false, true},
		{"ping", "", false, true},
		{"ack", "one", false, false},
		{"ack", "two", false, false},
		{"ack", "", false, false},
	}

	one := newTestService("one", t)
	two := newTestService("two", t)
	b1 := NewBroker()
	b2 := NewBroker()
	go func() {
		t.Fatal(b1.ListenOnConn(one.NewLocalConn()))
	}()
	go func() {
		t.Fatal(b2.ListenOnConn(two.NewLocalConn()))
	}()
	time.Sleep(10 * time.Millisecond)

	one.Publish(Subscribe("ping", "one"))
	two.Publish(CreateMessage("proto/subs", []subscription{
		{"ping", "two", nil},
		{"ping", "", nil},
	}))
	time.Sleep(10 * time.Millisecond)

	go func() {
		t.Fatal(b2.ListenOnBridge(b1.NewLocalConn()))
	}()
	time.Sleep(10 * time.Millisecond)

	for i, test := range tests {
		one.Reset()
		two.Reset()
		one.Publish(Message{
			Id:          GenerateId(),
			Action:      test.action,
			Destination: test.device,
		})
		time.Sleep(10 * time.Millisecond)

		if test.oneShould && !one.Fired() {
			t.Error(i, "one did not fire", test)
		}
		if !test.oneShould && one.Fired() {
			t.Error(i, "one should not fire", test)
		}
		if test.twoShould && !two.Fired() {
			t.Error(i, "two did not fire", test)
		}
		if !test.twoShould && two.Fired() {
			t.Error(i, "two should not fire", test)
		}
	}
}

func TestBrokerGateway(t *testing.T) {
	tests := []brokerMultiEp{
		{"ping", "one", true, false},
		{"ping", "two", false, true},
		{"ping", "", false, true},
		{"ack", "one", false, false},
		{"ack", "two", false, false},
		{"ack", "", false, false},
	}

	one := newTestService("one", t)
	two := newTestService("two", t)
	b1 := NewBroker()
	b2 := NewBroker()
	go func() {
		t.Fatal(b1.ListenOnConn(one.NewLocalConn()))
	}()
	go func() {
		t.Fatal(b2.ListenOnConn(two.NewLocalConn()))
	}()
	time.Sleep(10 * time.Millisecond)

	one.Publish(Subscribe("ping", "one"))
	two.Publish(CreateMessage("proto/subs", []subscription{
		{"ping", "two", nil},
		{"ping", "", nil},
	}))
	time.Sleep(10 * time.Millisecond)

	go func() {
		t.Fatal(b2.ListenOnGateway(b1.NewLocalConn()))
	}()
	time.Sleep(10 * time.Millisecond)

	for i, test := range tests {
		one.Reset()
		two.Reset()
		one.Publish(Message{
			Id:          GenerateId(),
			Action:      test.action,
			Destination: test.device,
		})
		time.Sleep(10 * time.Millisecond)

		if test.oneShould && !one.Fired() {
			t.Error(i, "one did not fire", test)
		}
		if !test.oneShould && one.Fired() {
			t.Error(i, "one should not fire", test)
		}
		if test.twoShould && !two.Fired() {
			t.Error(i, "two did not fire", test)
		}
		if !test.twoShould && two.Fired() {
			t.Error(i, "two should not fire", test)
		}
	}
}

func BenchmarkBrokerSingle(b *testing.B) {
	br := NewBroker()
	c := br.NewLocalConn()
	c.Write(Subscribe("testaction", ""))

	msg := Message{
		Version: VERSION,
		Action:  "testaction",
		Source:  "me",
		Text:    "this text ist a bit longer to test things.",
	}
	msg.EncodePayload(map[string]interface{}{
		"a_value":   3,
		"some_text": "additional info",
		"nested_magic": []string{
			"yeah",
			"right",
			"this is okay",
		},
	})

	for n := 0; n < b.N; n++ {
		msg2 := msg
		msg2.Id = GenerateId()
		c.Write(msg2)
		c.Read()
	}
}

func BenchmarkBrokerSingleParallel(b *testing.B) {
	br := NewBroker()
	c := br.NewLocalConn()
	c.Write(Subscribe("testaction", ""))

	msg := Message{
		Version: VERSION,
		Action:  "testaction",
		Source:  "me",
		Text:    "this text ist a bit longer to test things.",
	}
	msg.EncodePayload(map[string]interface{}{
		"a_value":   3,
		"some_text": "additional info",
		"nested_magic": []string{
			"yeah",
			"right",
			"this is okay",
		},
	})

	go func() {
		for n := 0; n < b.N; n++ {
			msg2 := msg
			msg2.Id = GenerateId()
			c.Write(msg2)
		}
	}()
	for n := 0; n < b.N; n++ {
		c.Read()
	}
}

func BenchmarkBrokerSingleNet(b *testing.B) {
	cfg := &NetConfig{
		Address: "tcp://127.0.0.1:23401",
	}
	br := NewBroker()
	go br.Listen(cfg)
	time.Sleep(10 * time.Millisecond)
	c, err := Dial(cfg)
	if err != nil {
		b.Fatal(err)
	}
	msg := Subscribe("testaction", "")
	msg.Source = "me"
	if err := c.Write(msg); err != nil {
		b.Fatal(err)
	}

	msg = Message{
		Id:      GenerateId(),
		Version: VERSION,
		Action:  "testaction",
		Source:  "me",
		Text:    "this text ist a bit longer to test things.",
	}
	msg.EncodePayload(map[string]interface{}{
		"a_value":   3,
		"some_text": "additional info",
		"nested_magic": []string{
			"yeah",
			"right",
			"this is okay",
		},
	})

	go func() {
		for n := 0; n < b.N; n++ {
			msg2 := msg
			msg2.Id = GenerateId()
			if err := c.Write(msg2); err != nil {
				b.Fatal(err)
			}
		}
	}()
	for n := 0; n < b.N; n++ {
		if _, err := c.Read(); err != nil {
			b.Fatal(err)
		}
	}
}
