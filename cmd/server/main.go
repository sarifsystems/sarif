// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// A simple server that can host different stark services.
//
// The module loading hides a few implementation details, so for a better
// introduction, look at cmd/starkping (it works serverless)
package main

import (
	"flag"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/log"

	_ "github.com/go-sql-driver/mysql"

	_ "github.com/xconstruct/stark/services/events"
	_ "github.com/xconstruct/stark/services/hostscan"
	_ "github.com/xconstruct/stark/services/know"
	_ "github.com/xconstruct/stark/services/lastfm"
	_ "github.com/xconstruct/stark/services/location"
	_ "github.com/xconstruct/stark/services/natural"
	_ "github.com/xconstruct/stark/services/router"
	_ "github.com/xconstruct/stark/services/scheduler"
	_ "github.com/xconstruct/stark/web"
	_ "github.com/xconstruct/stark/xmpp"
)

var verbose = flag.Bool("v", false, "verbose debug output")

type Config struct {
	EnabledModules []string
}

func main() {
	flag.Parse()

	app := core.NewApp("stark")
	app.Must(app.Init())
	defer app.Close()

	if *verbose {
		app.Log.SetLevel(log.LevelDebug)
	}

	// Default configuration
	cfg := Config{
		EnabledModules: []string{
			"events",
			"location",
			"natural",
			"scheduler",
			"web",
		},
	}
	// Load configuration from file
	app.Must(app.Config.Get("server", &cfg))

	// Enable each module listed in the config
	for _, module := range cfg.EnabledModules {
		app.Must(app.EnableModule(module))
	}

	app.WaitUntilInterrupt()
}
