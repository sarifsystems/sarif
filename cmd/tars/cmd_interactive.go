// Copyright (C) 2015 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"io"
	"log"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/shiena/ansicolor"
)

var profile = flag.Bool("profile", false, "interactive: print elapsed time for requests")

func (app *App) Interactive() {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:      color.BlueString("say » "),
		HistoryFile: app.Config.HistoryFile,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()
	app.Log.SetOutput(rl.Stderr())
	log.SetOutput(rl.Stderr())
	color.Output = ansicolor.NewAnsiColorWriter(rl.Stderr())

	pings := make(map[string]time.Time)

	// Subscribe to all replies and print them to stdout
	app.Client.Subscribe("", "self", func(msg sarif.Message) {
		text := msg.Text
		if text == "" {
			text = msg.Action + " from " + msg.Source
		}
		if msg.IsAction("err") {
			text = color.RedString(text)
		}

		if sent, ok := pings[msg.CorrId]; ok {
			text += color.YellowString("[%.1fms]", time.Since(sent).Seconds()*1e3)
		}
		log.Println(color.GreenString(" « ") + strings.Replace(text, "\n", "\n   ", -1))
	})

	// Interactive mode sends all lines from stdin.
	for {
		line, err := rl.Readline()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Fatal(err)
		}
		if len(line) == 0 {
			continue
		}

		// Publish natural message
		msg := sarif.Message{
			Id:     sarif.GenerateId(),
			Action: "natural/handle",
			Text:   line,
		}
		if *profile {
			pings[msg.Id] = time.Now()
		}
		app.Client.Publish(msg)
	}
}
