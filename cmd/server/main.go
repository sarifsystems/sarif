// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/log"

	_ "github.com/go-sql-driver/mysql"

	_ "github.com/xconstruct/stark/services/hostscan"
	_ "github.com/xconstruct/stark/services/location"
	_ "github.com/xconstruct/stark/services/router"
	_ "github.com/xconstruct/stark/web"
	_ "github.com/xconstruct/stark/xmpp"
)

var verbose = flag.Bool("v", false, "verbose debug output")

type Config struct {
	EnabledModules []string
}

func main() {
	flag.Parse()

	app, err := core.NewApp("stark")
	app.Must(err)
	defer app.Close()

	if *verbose {
		app.Log.SetLevel(log.LevelDebug)
	}

	cfg := Config{
		EnabledModules: []string{"web"},
	}
	app.Must(app.Config.Get("server", &cfg))

	for _, module := range cfg.EnabledModules {
		app.Must(app.EnableModule(module))
	}

	app.WaitUntilInterrupt()
}
