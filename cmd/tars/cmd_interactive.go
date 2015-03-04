// Copyright (C) 2015 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"io"
	"log"
	"os"

	"github.com/xconstruct/stark/proto"
)

func (app *App) Interactive() {
	// Subscribe to all replies and print them to stdout
	app.Client.Subscribe("", "self", func(msg proto.Message) {
		text := msg.Text
		if text == "" {
			text = msg.Action + " from " + msg.Source
		}
		log.Println(text)
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
		app.Client.Publish(proto.Message{
			Action: "natural/handle",
			Text:   string(line),
		})
	}
}
