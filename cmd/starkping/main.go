// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// A simple stark client that pings the network every second and prints the
// results.
//
// Example: ./starkping tcp://localhost:23100
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/xconstruct/stark/proto"
)

func main() {
	addr := "tcp://localhost:23100"
	if len(os.Args) > 1 {
		addr = os.Args[1]
	}
	fmt.Println("connecting to", addr)

	// Dial into the stark network.
	conn, err := proto.Dial(&proto.NetConfig{
		Address: addr,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Setup our client.
	client := proto.NewClient("starkping", conn)
	pings := make(map[string]time.Time)

	// Subscribe to all acknowledgements to our pings
	// and print them.
	client.Subscribe("ack", "self", func(msg proto.Message) {
		if !msg.IsAction("ack") {
			return
		}
		sent, ok := pings[msg.CorrId]
		if !ok {
			return
		}

		fmt.Printf("%s from %s: time=%.1fms\n",
			msg.Action,
			msg.Source,
			time.Since(sent).Seconds()*1e3,
		)
	})

	// Every second, send a ping to all devices.
	for _ = range time.Tick(1 * time.Second) {
		// Create the ping message and publish it on the network
		msg := proto.Message{
			Id:     proto.GenerateId(),
			Action: "ping",
		}
		if err := client.Publish(msg); err != nil {
			log.Fatal(err)
		}
		pings[msg.Id] = time.Now()
	}
}
