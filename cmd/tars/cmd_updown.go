// Copyright (C) 2015 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/sarifsystems/sarif/sarif"
)

type ContentPayload struct {
	ContentType string `json:"content_type,omitempty"`
	Content     []byte `json:"content,omitempty"`
}

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
	pl := &ContentPayload{
		ContentType: http.DetectContentType(in),
		Content:     in,
	}
	if strings.HasPrefix(pl.ContentType, "text/plain;") {
		msg.Text = string(pl.Content)
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
