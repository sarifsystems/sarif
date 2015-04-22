// Copyright (C) 2015 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, Usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	app := New()
	app.Run()
}

type App struct {
	*core.App
	Client   *proto.Client
	Commands []Command
}

type Command struct {
	Name string
	Do   func()
	Help string
}

func New() *App {
	log.SetFlags(0)
	app := &App{
		App:    core.NewApp("stark", "client"),
		Client: proto.NewClient("tars/" + proto.GenerateId()),
	}
	app.Init()
	app.Must(app.Client.Connect(app.Dial()))
	app.Client.OnConnectionLost(func(err error) {
		app.Log.Fatalln("connection lost:", err)
	})

	app.Commands = []Command{
		{"help", app.Help, ""},
		{"interactive", app.Interactive, ""},
		{"location_import", app.LocationImport, ""},
		{"cat", app.Cat, usageCat},
		{"down", app.Down, ""},
		{"up", app.Up, ""},
	}

	return app
}

func (app *App) Run() {
	defer app.Close()
	cmd := "interactive"
	if flag.NArg() > 0 {
		cmd = flag.Arg(0)
	}

	for _, c := range app.Commands {
		if c.Name == cmd {
			c.Do()
			return
		}
	}

	// Default
	app.SingleRequest()
}
