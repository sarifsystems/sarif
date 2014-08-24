package mux

import (
	"testing"
	"time"

	"github.com/xconstruct/stark/proto"
)

type multiEp struct {
	action    string
	device    string
	oneShould bool
	twoShould bool
}

func TestTransportMuxMultiple(t *testing.T) {
	tests := []multiEp{
		{"ping", "one", true, false},
		{"ping", "two", false, true},
		{"ping", "", false, true},
		{"ack", "one", false, false},
		{"ack", "two", false, false},
		{"ack", "", false, false},
	}

	mux := NewTransportMux()
	mux.RegisterPublisher(func(msg proto.Message) error {
		return nil
	})
	oneFired, twoFired := false, false

	epOne := mux.NewEndpoint()
	epOne.RegisterHandler(func(msg proto.Message) {
		oneFired = true
	})
	epOne.Publish(proto.Message{
		Action: "proto/sub",
		Payload: map[string]interface{}{
			"action": "ping",
			"device": "one",
		},
	})

	epTwo := mux.NewEndpoint()
	epTwo.RegisterHandler(func(msg proto.Message) {
		twoFired = true
	})
	epTwo.Publish(proto.Message{
		Action: "proto/sub",
		Payload: map[string]interface{}{
			"action": "ping",
			"device": "two",
		},
	})
	epTwo.Publish(proto.Message{
		Action: "proto/sub",
		Payload: map[string]interface{}{
			"action": "ping",
			"device": "",
		},
	})

	for _, test := range tests {
		oneFired, twoFired = false, false
		mux.Handle(proto.Message{
			Action: test.action,
			Device: test.device,
		})
		time.Sleep(time.Millisecond)
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
