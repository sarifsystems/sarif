// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// A simple stark client that sends to the network and listens for a response.
package proto_test

import (
	"fmt"
	"time"

	"github.com/xconstruct/stark/proto"
)

func ExampleClient() {
	recv := make(chan bool)

	// Connect to MQTT network
	conn, err := proto.DialMqtt(proto.MqttConfig{
		Server: "tcp://test.mosquitto.org:1883",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	// Setup client and listen for "testaction"
	client := proto.NewClient("mytestdevice", conn)
	client.Subscribe("testaction", "", func(msg proto.Message) {
		fmt.Printf("received %q from %q\n", msg.Action, msg.Source)

		var payload interface{}
		msg.DecodePayload(&payload)
		fmt.Println("payload:", payload)
		recv <- true
	})

	// Publish a "testaction" message
	client.Publish(proto.CreateMessage("testaction", "weee, a test message!"))

	// Block until message received or timeout
	select {
	case <-recv:
	case <-time.After(time.Second):
	}

	// Output:
	// received "testaction" from "mytestdevice"
	// payload: weee, a test message!
}
