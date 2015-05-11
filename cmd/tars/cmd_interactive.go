// Copyright (C) 2015 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"time"

	"github.com/xconstruct/stark/proto"
)

var profile = flag.Bool("profile", false, "interactive: print elapsed time for requests")

func (app *App) Interactive() {
	pings := make(map[string]time.Time)

	// Subscribe to all replies and print them to stdout
	app.Client.Subscribe("", "self", func(msg proto.Message) {
		text := msg.Text
		if text == "" {
			text = msg.Action + " from " + msg.Source
		}

		if sent, ok := pings[msg.CorrId]; ok {
			log.Printf("%s [%.1fms]\n", text, time.Since(sent).Seconds()*1e3)
		} else {
			log.Println(text)
		}
	})

	// Interactive mode sends all lines from stdin.
	in := bufio.NewReader(os.Stdin)
	for {
		line, _, err := in.ReadLine()
		if err != nil {
			if err == io.EOF {
				os.Exit(0)
			}
			log.Fatal(err)
		}
		if len(line) == 0 {
			continue
		}

		// Publish natural message
		msg := proto.Message{
			Id:     proto.GenerateId(),
			Action: "natural/handle",
			Text:   string(line),
		}
		if *profile {
			pings[msg.Id] = time.Now()
		}
		app.Client.Publish(msg)
	}
}
