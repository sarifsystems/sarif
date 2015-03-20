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
	"github.com/xconstruct/stark/core/server"
	"github.com/xconstruct/stark/services/commands"
	"github.com/xconstruct/stark/services/events"
	"github.com/xconstruct/stark/services/hostscan"
	"github.com/xconstruct/stark/services/know"
	"github.com/xconstruct/stark/services/lastfm"
	"github.com/xconstruct/stark/services/location"
	"github.com/xconstruct/stark/services/luascripts"
	"github.com/xconstruct/stark/services/meals"
	"github.com/xconstruct/stark/services/mood"
	"github.com/xconstruct/stark/services/natural"
	"github.com/xconstruct/stark/services/router"
	"github.com/xconstruct/stark/services/scheduler"
	"github.com/xconstruct/stark/services/statenet"
	"github.com/xconstruct/stark/services/states"
	"github.com/xconstruct/stark/services/store"
	"github.com/xconstruct/stark/services/trigger"
	"github.com/xconstruct/stark/services/web"
	"github.com/xconstruct/stark/services/xmpp"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

func main() {
	srv := server.New("stark", "server")

	srv.RegisterModule(commands.Module)
	srv.RegisterModule(events.Module)
	srv.RegisterModule(hostscan.Module)
	srv.RegisterModule(know.Module)
	srv.RegisterModule(lastfm.Module)
	srv.RegisterModule(location.Module)
	srv.RegisterModule(luascripts.Module)
	srv.RegisterModule(meals.Module)
	srv.RegisterModule(mood.Module)
	srv.RegisterModule(natural.Module)
	srv.RegisterModule(router.Module)
	srv.RegisterModule(scheduler.Module)
	srv.RegisterModule(states.Module)
	srv.RegisterModule(statenet.Module)
	srv.RegisterModule(store.Module)
	srv.RegisterModule(trigger.Module)
	srv.RegisterModule(web.Module)
	srv.RegisterModule(xmpp.Module)

	// Default configuration
	srv.ServerConfig = server.Config{
		EnabledModules: []string{
			"commands",
			"events",
			"know",
			"location",
			"meals",
			"mood",
			"natural",
			"scheduler",
			"states",
			"store",
			"trigger",
			"web",
		},
	}

	srv.Run()
}
