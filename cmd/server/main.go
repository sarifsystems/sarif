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
	"github.com/xconstruct/stark/core"

	_ "github.com/go-sql-driver/mysql"

	_ "github.com/xconstruct/stark/services/events"
	_ "github.com/xconstruct/stark/services/hostscan"
	_ "github.com/xconstruct/stark/services/know"
	_ "github.com/xconstruct/stark/services/lastfm"
	_ "github.com/xconstruct/stark/services/location"
	_ "github.com/xconstruct/stark/services/natural"
	_ "github.com/xconstruct/stark/services/router"
	_ "github.com/xconstruct/stark/services/scheduler"
	_ "github.com/xconstruct/stark/services/store"
	_ "github.com/xconstruct/stark/services/web"
	_ "github.com/xconstruct/stark/services/xmpp"
)

type Config struct {
	EnabledModules []string
}

func main() {
	app := core.NewApp("stark")
	app.Must(app.Init())
	defer app.Close()

	// Default configuration
	cfg := Config{
		EnabledModules: []string{
			"events",
			"know",
			"location",
			"natural",
			"scheduler",
			"store",
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
