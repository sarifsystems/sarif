// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// A simple server that can host different sarif services.
//
// The module loading hides a few implementation details, so for a better
// introduction, look at cmd/sarifping (it works serverless)
package main

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/sarifsystems/sarif/core/server"
	"github.com/sarifsystems/sarif/services/commands"
	"github.com/sarifsystems/sarif/services/events"
	"github.com/sarifsystems/sarif/services/hostscan"
	"github.com/sarifsystems/sarif/services/know"
	"github.com/sarifsystems/sarif/services/lastfm"
	"github.com/sarifsystems/sarif/services/location"
	"github.com/sarifsystems/sarif/services/logger"
	"github.com/sarifsystems/sarif/services/luascripts"
	"github.com/sarifsystems/sarif/services/meals"
	"github.com/sarifsystems/sarif/services/mock"
	"github.com/sarifsystems/sarif/services/natural"
	"github.com/sarifsystems/sarif/services/nlparser"
	"github.com/sarifsystems/sarif/services/nlquery"
	"github.com/sarifsystems/sarif/services/reasoner"
	"github.com/sarifsystems/sarif/services/scheduler"
	"github.com/sarifsystems/sarif/services/store"
	_ "github.com/sarifsystems/sarif/services/store/bolt"
	"github.com/sarifsystems/sarif/services/vdir"
	"github.com/sarifsystems/sarif/services/web"
	"github.com/sarifsystems/sarif/services/xmpp"
)

func main() {
	srv := server.New("sarif", "sarifd")

	srv.RegisterModule(commands.Module)
	srv.RegisterModule(events.Module)
	srv.RegisterModule(hostscan.Module)
	srv.RegisterModule(know.Module)
	srv.RegisterModule(lastfm.Module)
	srv.RegisterModule(logger.Module)
	srv.RegisterModule(location.Module)
	srv.RegisterModule(luascripts.Module)
	srv.RegisterModule(meals.Module)
	srv.RegisterModule(mock.Module)
	srv.RegisterModule(natural.Module)
	srv.RegisterModule(nlparser.Module)
	srv.RegisterModule(nlquery.Module)
	srv.RegisterModule(reasoner.Module)
	srv.RegisterModule(scheduler.Module)
	srv.RegisterModule(store.Module)
	srv.RegisterModule(vdir.Module)
	srv.RegisterModule(web.Module)
	srv.RegisterModule(xmpp.Module)

	// Default configuration
	srv.ServerConfig.BaseModules = []string{
		"store",
	}
	srv.ServerConfig.EnabledModules = []string{
		"commands",
		"events",
		"know",
		"location",
		"logger",
		"mock",
		"meals",
		"natural",
		"nlparser",
		"nlquery",
		"reasoner",
		"scheduler",
		"vdir",
		"web",
	}

	srv.Run()
}
