// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// A simple stark client that pings the network every second and prints the
// results.
//
// You probably want to start it with the -v (verbose) flag to get a feel
// for the protocol.
package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/log"
	"github.com/xconstruct/stark/proto"
)

var verbose = flag.Bool("v", false, "verbose debug output")

// PingService can send ping messages and measures the elapsed time until
// a response is received.
type PingService struct {
	pings  map[string]time.Time
	client *proto.Client
}

// NewPingService creates a PingService that communicates on the supplied
// stark protocol connection endpoint.
func NewPingService(ep proto.Endpoint) *PingService {
	// We create a new client on this connection with an unique name.

	// While the raw endpoint is good enough for sending/receiving pure messages,
	// the client provides lots of helpful abstractions and handles a few
	// implementation details.
	name := "starkping-" + proto.GenerateId()
	client := proto.NewClient(name, ep)

	return &PingService{
		make(map[string]time.Time),
		client,
	}
}

// Ping sends a new ping message to the specified device in the stark network.
// If device is an empty string, sends it to the whole network.
// By spec, every client is normally bound to respond with an "ack".
func (s *PingService) Ping(device string) error {
	// Generate an unique ID for our message and store the time we sent it.
	id := proto.GenerateId()
	s.pings[id] = time.Now()

	// Create the ping message and publish it on the network
	msg := proto.Message{
		Id:          id,
		Action:      "ping",
		Destination: device,
	}
	return s.client.Publish(msg)
}

// Enable starts the service by subscribing to the right stark messages.
func (s *PingService) Enable() error {
	// Listen for messages with action "ack" that are send directly to us
	// and pass them to Handle()
	return s.client.Subscribe("ack", "self", s.Handle)
}

// Handle processes an incoming stark message.
func (s *PingService) Handle(msg proto.Message) {
	// We want to only handle acknowledgements to our pings here
	if !msg.IsAction("ack") {
		return
	}
	// Does the ack reference a previous ping message in its correlation id?
	sent, ok := s.pings[msg.CorrId]
	if !ok {
		return
	}

	// Print the received message and the elapsed time since the ping.
	fmt.Printf("%s from %s: time=%.1fms\n",
		msg.Action,
		msg.Source,
		time.Since(sent).Seconds()*1e3,
	)
}

func main() {
	flag.Parse()

	// App simply helps to read our global configuration file and sets up the
	// MQTT connection to the network. It is not strictly necessary for own
	// services.
	app := core.NewApp("stark")
	app.Must(app.Init())
	defer app.Close()
	if *verbose {
		app.Log.SetLevel(log.LevelDebug)
	}
	ctx := app.NewContext()

	// Enable our own stark Service.
	srv := NewPingService(ctx.Proto)
	ctx.Must(srv.Enable())

	// Every second, send a ping to all devices.
	for _ = range time.Tick(1 * time.Second) {
		ctx.Must(srv.Ping(""))
	}
}
