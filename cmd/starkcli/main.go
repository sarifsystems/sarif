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
	"github.com/xconstruct/stark/log"
	"github.com/xconstruct/stark/proto"
)

var (
	verbose = flag.Bool("v", false, "verbose debug output")
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
	app := core.NewApp("stark")
	if *verbose {
		app.Log.SetLevel(log.LevelDebug)
	} else {
		app.Log.SetLevel(log.LevelWarn)
	}
	app.Must(app.Init())
	defer app.Close()
	ctx := app.NewContext()

	// Connect to network.
	name := "starkcli-" + proto.GenerateId()
	client := proto.NewClient(name, ctx.Proto)

	// Subscribe to all replies and print them to stdout
	client.Subscribe("", "self", func(msg proto.Message) {
		text := msg.PayloadGetString("text")
		if text == "" {
			text = msg.Action + " from " + msg.Source
		}
		fmt.Println(text)
		// Non-interactive? Exit after receiving message
		if !isInteractive {
			os.Exit(0)
		}
	})

	if isInteractive {
		// Interactive mode sends all lines from stdin.
		in := bufio.NewReader(os.Stdin)
		for {
			line, _, err := in.ReadLine()
			if err != nil {
				if err == io.EOF {
					os.Exit(0)
				}
				ctx.Log.Fatal(err)
			}
			if string(line) == "" {
				continue
			}

			// Publish natural message
			client.Publish(proto.Message{
				Action: "natural/handle",
				Payload: map[string]interface{}{
					"text": string(line),
				},
			})
		}
	} else {
		// Non-interactive mode publishes arguments and waits for response.
		client.Publish(proto.Message{
			Action: "natural/handle",
			Payload: map[string]interface{}{
				"text": strings.Join(flag.Args(), " "),
			},
		})
		select {}
	}
}
