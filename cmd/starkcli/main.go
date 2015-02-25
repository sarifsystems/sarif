// Copyright (C) 2015 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Natural commandline interface to the stark network.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
)

const Usage = `Usage: starkcli [OPTION]... [MESSAGE]...
Natural commandline interface to the stark .
Publishes a natural message on the stark network, prints the response
and returns.

If invoked with no arguments, it starts an interactive session, where
each line from stdin published as a natural message.

Options:

`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, Usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	isInteractive := flag.NArg() == 0

	// Setup app and read config.
	app := core.NewApp("stark", "client")
	app.Init()
	defer app.Close()

	// Connect to network.
	name := "starkcli-" + proto.GenerateId()
	client := proto.NewClient(name)
	client.Connect(app.Dial())

	// Subscribe to all replies and print them to stdout
	client.Subscribe("", "self", func(msg proto.Message) {
		text := msg.Text
		if text == "" {
			text = msg.Action + " from " + msg.Source
		}
		fmt.Println(text)
		// Non-interactive? Exit after receiving message
		if !isInteractive {
			os.Exit(0)
		}
	})

	go func() {
		if isInteractive {
			// Interactive mode sends all lines from stdin.
			in := bufio.NewReader(os.Stdin)
			for {
				line, _, err := in.ReadLine()
				if err != nil {
					if err == io.EOF {
						os.Exit(0)
					}
					app.Log.Fatal(err)
				}
				if string(line) == "" {
					continue
				}

				// Publish natural message
				client.Publish(proto.Message{
					Action: "natural/handle",
					Text:   string(line),
				})
			}
		} else {
			// Non-interactive mode publishes arguments and waits for response.
			client.Publish(proto.Message{
				Action: "natural/handle",
				Text:   strings.Join(flag.Args(), " "),
			})

		}
	}()
	core.WaitUntilInterrupt()
}
