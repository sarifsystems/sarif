// Copyright (C) 2015 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/xconstruct/stark/proto"
)

var (
	waitNum = flag.Int("n", -1, "cat: wait for X replies before exiting (-1 for indefinitely)")
	timeout = flag.Duration("t", 1*time.Second, "cat: wait for X duration before exiting (-1s for indefinitely)")
)

const usageCat = `Usage: tars [OPTION}... cat [ACTION]...
Publish and subscribe to messages in the stark network.
Accepts JSON-encoded messages on stdin and prints replies on stdout.

By default, the client only subscribes to messages directed at itself.
You can also provide actions to subscribe to as additional arguments,
optionally in the format "action:device" (e.g. "ping", "ping:self",
"ping:mydevice123").

    Example: Publish a message and wait 2 seconds for replies
        echo '{"action":"ping"}' | tars cat

    Example: Listen for the next global ping message
        tars -n 1 -t -1s cat ping
`

func (app *App) Cat() {
	received := make(chan bool, 10)

	// Handle replies: print them as readable JSON.
	handle := func(msg proto.Message) {
		raw, err := json.MarshalIndent(msg, "", "    ")
		app.Must(err)
		log.Println(string(raw))
		*waitNum -= 1
		if *waitNum == 0 {
			received <- true
		}
	}

	// Subscribe to all topics we're interested in.
	if flag.NArg() <= 1 {
		app.Client.Subscribe("", "self", handle)
	}
	for i, action := range flag.Args() {
		if i == 0 {
			continue
		}
		parts := strings.Split(action, ":")
		if len(parts) > 1 {
			app.Client.Subscribe(parts[0], parts[1], handle)
		} else {
			app.Client.Subscribe(parts[0], "", handle)
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
			app.Must(app.Client.Publish(msg))
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
			app.Log.Debugln("All messages received, exiting ...")
			break OUTER
		case <-timer:
			app.Log.Debugln("Timeout, exiting ...")
			break OUTER
		}
	}
}
