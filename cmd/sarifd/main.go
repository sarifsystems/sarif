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
	"github.com/sarifsystems/sarif/services/auth"
	"github.com/sarifsystems/sarif/services/commands"
	"github.com/sarifsystems/sarif/services/events"
	"github.com/sarifsystems/sarif/services/hostscan"
	"github.com/sarifsystems/sarif/services/js"
	"github.com/sarifsystems/sarif/services/know"
	"github.com/sarifsystems/sarif/services/location"
	"github.com/sarifsystems/sarif/services/logger"
	"github.com/sarifsystems/sarif/services/lua"
	"github.com/sarifsystems/sarif/services/mock"
	"github.com/sarifsystems/sarif/services/natural"
	"github.com/sarifsystems/sarif/services/nlparser"
	"github.com/sarifsystems/sarif/services/nlquery"
	"github.com/sarifsystems/sarif/services/pushgateway"
	"github.com/sarifsystems/sarif/services/scheduler"
	"github.com/sarifsystems/sarif/services/scrobbler"
	"github.com/sarifsystems/sarif/services/spotify"
	"github.com/sarifsystems/sarif/services/store"
	_ "github.com/sarifsystems/sarif/services/store/bolt"
	_ "github.com/sarifsystems/sarif/services/store/es7"
	_ "github.com/sarifsystems/sarif/services/store/replicate"
	"github.com/sarifsystems/sarif/services/vdir"
	"github.com/sarifsystems/sarif/services/web"
	"github.com/sarifsystems/sarif/services/xmpp"
)

func main() {
	srv := server.New("sarif", "sarifd")

	srv.RegisterModule(auth.Module)
	srv.RegisterModule(commands.Module)
	srv.RegisterModule(events.Module)
	srv.RegisterModule(hostscan.Module)
	srv.RegisterModule(know.Module)
	srv.RegisterModule(logger.Module)
	srv.RegisterModule(location.Module)
	srv.RegisterModule(lua.Module)
	srv.RegisterModule(js.Module)
	srv.RegisterModule(mock.Module)
	srv.RegisterModule(natural.Module)
	srv.RegisterModule(nlparser.Module)
	srv.RegisterModule(nlquery.Module)
	srv.RegisterModule(pushgateway.Module)
	srv.RegisterModule(scheduler.Module)
	srv.RegisterModule(scrobbler.Module)
	srv.RegisterModule(spotify.Module)
	srv.RegisterModule(store.Module)
	srv.RegisterModule(vdir.Module)
	srv.RegisterModule(web.Module)
	srv.RegisterModule(xmpp.Module)

	// Default configuration
	srv.ServerConfig.BaseModules = []string{
		"store",
	}
	srv.ServerConfig.EnabledModules = []string{
		"auth",
		"commands",
		"events",
		"know",
		"location",
		"logger",
		"lua",
		"mock",
		"natural",
		"nlparser",
		"nlquery",
		"scheduler",
		"vdir",
		"web",
	}

	srv.Run()
}
