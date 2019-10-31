// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/sarifsystems/sarif/sarif"
)

func (app *App) LocationImport() {
	client := app.NewClient()
	fname := flag.Arg(1)
	body, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}

	result, ok := <-client.Request(sarif.CreateMessage("location/import", map[string]string{
		"csv": string(body),
	}))
	if !ok {
		log.Fatal("Timeout while waiting for import confirmation.")
	}
	if result.IsAction("err") {
		log.Fatal(result.Text)
	}
	log.Println(result.Text)
}
