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
	"github.com/xconstruct/stark/services/events"
	"github.com/xconstruct/stark/services/hostscan"
	"github.com/xconstruct/stark/services/know"
	"github.com/xconstruct/stark/services/lastfm"
	"github.com/xconstruct/stark/services/location"
	"github.com/xconstruct/stark/services/mood"
	"github.com/xconstruct/stark/services/natural"
	"github.com/xconstruct/stark/services/router"
	"github.com/xconstruct/stark/services/scheduler"
	"github.com/xconstruct/stark/services/store"
	"github.com/xconstruct/stark/services/web"
	"github.com/xconstruct/stark/services/xmpp"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	EnabledModules []string
}

func main() {
	app := core.NewApp("stark")
	app.Must(app.Init())
	defer app.Close()

	app.RegisterModule(events.Module)
	app.RegisterModule(hostscan.Module)
	app.RegisterModule(know.Module)
	app.RegisterModule(lastfm.Module)
	app.RegisterModule(location.Module)
	app.RegisterModule(mood.Module)
	app.RegisterModule(natural.Module)
	app.RegisterModule(router.Module)
	app.RegisterModule(scheduler.Module)
	app.RegisterModule(store.Module)
	app.RegisterModule(web.Module)
	app.RegisterModule(xmpp.Module)

	// Default configuration
	cfg := Config{
		EnabledModules: []string{
			"events",
			"know",
			"location",
			"mood",
			"natural",
			"scheduler",
			"store",
			"web",
		},
	}
	// Load configuration from file
	app.Config.Get("server", &cfg)

	// Enable each module listed in the config
	for _, module := range cfg.EnabledModules {
		app.Must(app.EnableModule(module))
	}

	app.WaitUntilInterrupt()
}
