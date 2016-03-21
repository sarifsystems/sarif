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
	"github.com/xconstruct/stark/services/logger"
	"github.com/xconstruct/stark/services/luascripts"
	"github.com/xconstruct/stark/services/meals"
	"github.com/xconstruct/stark/services/natural"
	"github.com/xconstruct/stark/services/nlparser"
	"github.com/xconstruct/stark/services/reasoner"
	"github.com/xconstruct/stark/services/scheduler"
	"github.com/xconstruct/stark/services/store"
	"github.com/xconstruct/stark/services/vdir"
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
	srv.RegisterModule(logger.Module)
	srv.RegisterModule(location.Module)
	srv.RegisterModule(luascripts.Module)
	srv.RegisterModule(meals.Module)
	srv.RegisterModule(natural.Module)
	srv.RegisterModule(nlparser.Module)
	srv.RegisterModule(reasoner.Module)
	srv.RegisterModule(scheduler.Module)
	srv.RegisterModule(store.Module)
	srv.RegisterModule(vdir.Module)
	srv.RegisterModule(web.Module)
	srv.RegisterModule(xmpp.Module)

	// Default configuration
	srv.ServerConfig.EnabledModules = []string{
		"commands",
		"events",
		"know",
		"location",
		"logger",
		"meals",
		"natural",
		"nlparser",
		"reasoner",
		"scheduler",
		"states",
		"store",
		"trigger",
		"vdir",
		"web",
	}

	srv.Run()
}
