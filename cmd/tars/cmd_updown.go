// Copyright (C) 2015 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/sarifsystems/sarif/pkg/content"
	"github.com/sarifsystems/sarif/sarif"
)

func (app *App) Down() {
	if flag.NArg() <= 1 {
		app.Log.Fatal("Please specify an action to listen for.")
	}
	action := flag.Arg(1)

	msg, ok := <-app.Client.Request(sarif.Message{
		Action: action,
	})
	if !ok {
		app.Log.Fatal("No reply received.")
	}
	fmt.Println(strings.TrimSpace(msg.Text))
}

func (app *App) Up() {
	if flag.NArg() <= 1 {
		app.Log.Fatal("Please specify an action to send to.")
	}
	action := flag.Arg(1)

	in, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		app.Log.Fatal(err)
	}

	msg := sarif.Message{
		Action: action,
	}

	pl := ContentPayload{
		Content: content.PutData(in),
	}
	if strings.HasPrefix(pl.Content.Type, "text/") {
		msg.Text = string(in)
	}
	if err := msg.EncodePayload(pl); err != nil {
		app.Log.Fatal(err)
	}

	msg, ok := <-app.Client.Request(msg)
	if !ok {
		app.Log.Fatal("No reply received.")
	}
	fmt.Println(strings.TrimSpace(msg.Text))
}
