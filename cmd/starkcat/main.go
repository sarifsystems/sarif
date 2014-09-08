// Copyright (C) 2015 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// A simple commandline client that publishes messages and listens for replies.
// For use in bash scripts or similar. See usage below or via "starkcat -h".
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/log"
	"github.com/xconstruct/stark/proto"
)

const Usage = `Usage: starkcat [OPTION]... [ACTION]...
Publish and subscribe to messages in the stark network.
Accepts JSON-encoded messages on stdin and prints replies on stdout.

By default, the client only subscribes to messages directed at itself.
You can also provide actions to subscribe to as additional arguments,
optionally in the format "action:device" (e.g. "ping", "ping:self",
"ping:mydevice123").

    Example: Publish a message and wait 2 seconds for replies
        echo '{"action":"ping"}' | starkcat

    Example: Listen for the next global ping message
        starkcat -n 1 -t -1s ping

Options:

`

var (
	verbose = flag.Bool("v", false, "verbose debug output")
	waitNum = flag.Int("n", -1, "wait for X replies before exiting (-1 for indefinitely)")
	timeout = flag.Duration("t", 2*time.Second, "wait for X duration before exiting (-1s for indefinitely)")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, Usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	// Setup app and read config.
	app, err := core.NewApp("stark")
	app.Must(err)
	defer app.Close()
	if *verbose {
		app.Log.SetLevel(log.LevelDebug)
	}
	ctx := app.NewContext()

	// Connect to network.
	name := "starkcat-" + proto.GenerateId()
	client := proto.NewClient(name, ctx.Proto)

	received := make(chan bool, 10)

	// Handle replies: print them as readable JSON.
	handle := func(msg proto.Message) {
		raw, err := json.MarshalIndent(msg, "", "    ")
		app.Must(err)
		fmt.Println(string(raw))
		*waitNum -= 1
		if *waitNum == 0 {
			received <- true
		}
	}

	// Subscribe to all topics we're interested in.
	if flag.NArg() == 0 {
		client.Subscribe("", "self", handle)
	}
	for _, action := range flag.Args() {
		parts := strings.Split(action, ":")
		if len(parts) > 1 {
			client.Subscribe(parts[0], parts[1], handle)
		} else {
			client.Subscribe(parts[0], "", handle)
		}
	}

	// Read stdin as json messages and publish to the network
	go func() {
		dec := json.NewDecoder(os.Stdin)
		for {
			var msg proto.Message
			if err := dec.Decode(&msg); err != nil {
				if err == io.EOF {
					break
				}
				app.Must(err)
			}
			app.Must(client.Publish(msg))
		}
	}()

	// Can we return immediately?
	if *waitNum == 0 || *timeout == 0 {
		return
	}

	// Set exit timer
	timer := make(<-chan time.Time)
	if *timeout > 0 {
		timer = time.After(*timeout)
	}

	// Wait until a exit condition is met.
OUTER:
	for {
		select {
		case <-received:
			ctx.Log.Infoln("All messages received, exiting ...")
			break OUTER
		case <-timer:
			ctx.Log.Infof("Timeout, exiting ...")
			break OUTER
		}
	}
}
