// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// A simple sarif client that sends to the network and listens for a response.
package sfproto_test

import (
	"fmt"
	"time"

	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/transports/sfproto"
)

type MyPayload struct {
	Value string
}

func ExampleClient() {
	recv := make(chan bool)

	// Hosting a broker
	broker := sfproto.NewBroker()
	go broker.Listen(&sfproto.NetConfig{
		Address: "tcp://localhost:5698",
		Auth:    sfproto.AuthNone,
	})
	time.Sleep(10 * time.Millisecond)

	// Setup client and listen for "testaction"
	client := sarif.NewClient(sarif.ClientInfo{
		Name: "mytestdevice",
	})
	conn, err := sfproto.Dial(&sfproto.NetConfig{
		Address: "tcp://localhost:5698",
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := client.Connect(conn); err != nil {
		fmt.Println(err)
		return
	}
	client.Subscribe("testaction", "", func(msg sarif.Message) {
		fmt.Printf("received %q from %q\n", msg.Action, msg.Source)

		var payload MyPayload
		msg.DecodePayload(&payload)
		fmt.Println("payload:", payload.Value)
		recv <- true
	})

	// Publish a "testaction" message
	client.Publish(sarif.CreateMessage("testaction", MyPayload{"weee, a test message!"}))

	// Block until message received or timeout
	select {
	case <-recv:
	case <-time.After(time.Second):
	}

	// Output:
	// received "testaction" from "mytestdevice"
	// payload: weee, a test message!
}
