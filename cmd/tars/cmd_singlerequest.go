// Copyright (C) 2015 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"strings"

	"github.com/xconstruct/stark/proto"
)

func (app *App) SingleRequest() {
	text := strings.Join(flag.Args(), " ")
	msg, ok := <-app.Client.Request(proto.Message{
		Action: "natural/handle",
		Text:   text,
	})

	if !ok {
		log.Fatal("No response received.")
	}
	if msg.Text != "" {
		log.Println(msg.Text)
	} else {
		log.Println(msg.Action + " from " + msg.Source)
	}
}
