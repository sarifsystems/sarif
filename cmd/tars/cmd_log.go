// Copyright (C) 2019 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"log"

	"github.com/sarifsystems/sarif/core"
	"github.com/sarifsystems/sarif/sarif"
)

func (app *App) CmdLog() {
	client := app.NewClient()
	app.Must(client.Subscribe("", "", func(msg sarif.Message) {
		body, err := json.Marshal(msg)
		app.Must(err)
		log.Println(string(body))
	}))

	core.WaitUntilInterrupt()
}
