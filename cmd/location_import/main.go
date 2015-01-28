// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
)

func main() {
	flag.Parse()
	app := core.NewApp("stark", "client")
	app.Init()

	if flag.NArg() == 0 {
		fmt.Println("expected csv file to import as argument")
		os.Exit(1)
	}

	fname := flag.Arg(0)
	body, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}

	conn := app.Dial()
	// Connect to network.
	name := "location_import-" + proto.GenerateId()
	client := proto.NewClient(name, conn)
	client.Subscribe("", "self", func(msg proto.Message) {})

	result, ok := <-client.Request(proto.CreateMessage("location/import", map[string]string{
		"csv": string(body),
	}))
	if !ok {
		log.Fatal("Timeout while waiting for import confirmation.")
	}
	if result.IsAction("err") {
		log.Fatal(result.Text)
	}
	fmt.Println(result.Text)
}
