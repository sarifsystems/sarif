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
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/core/server"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services/commands"
	"github.com/xconstruct/stark/services/events"
	"github.com/xconstruct/stark/services/hostscan"
	"github.com/xconstruct/stark/services/know"
	"github.com/xconstruct/stark/services/lastfm"
	"github.com/xconstruct/stark/services/location"
	"github.com/xconstruct/stark/services/luascripts"
	"github.com/xconstruct/stark/services/mood"
	"github.com/xconstruct/stark/services/natural"
	"github.com/xconstruct/stark/services/router"
	"github.com/xconstruct/stark/services/scheduler"
	"github.com/xconstruct/stark/services/store"
	"github.com/xconstruct/stark/services/timeseries"
	"github.com/xconstruct/stark/services/trigger"
	"github.com/xconstruct/stark/services/web"
	"github.com/xconstruct/stark/services/xmpp"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

type Config struct {
	Listen         []*proto.NetConfig
	Bridges        []*proto.NetConfig
	EnabledModules []string
}

func main() {
	app := server.Init("stark", "server")
	app.InitDatabase()
	app.InitBroker()
	defer app.Close()

	app.RegisterModule(commands.Module)
	app.RegisterModule(events.Module)
	app.RegisterModule(hostscan.Module)
	app.RegisterModule(know.Module)
	app.RegisterModule(lastfm.Module)
	app.RegisterModule(location.Module)
	app.RegisterModule(luascripts.Module)
	app.RegisterModule(mood.Module)
	app.RegisterModule(natural.Module)
	app.RegisterModule(router.Module)
	app.RegisterModule(scheduler.Module)
	app.RegisterModule(store.Module)
	app.RegisterModule(timeseries.Module)
	app.RegisterModule(trigger.Module)
	app.RegisterModule(web.Module)
	app.RegisterModule(xmpp.Module)

	// Default configuration
	cfg := Config{
		EnabledModules: []string{
			"commands",
			"events",
			"know",
			"location",
			"mood",
			"natural",
			"scheduler",
			"store",
			"trigger",
			"web",
		},
	}
	// Load configuration from file
	app.Config.Get("server", &cfg)

	if len(cfg.Listen) == 0 {
		cfg.Listen = append(cfg.Listen, &proto.NetConfig{
			Address: "tcp://localhost:23100",
		})
		app.Config.Set("server", &cfg)
	}

	// Listen on connections
	for _, cfg := range cfg.Listen {
		go func(cfg *proto.NetConfig) {
			app.Log.Infoln("[server] listening on", cfg.Address)
			app.Must(app.Broker.Listen(cfg))
		}(cfg)
	}

	// Setup bridges
	for _, cfg := range cfg.Bridges {
		go func(cfg *proto.NetConfig) {
			for {
				app.Log.Infoln("[server] bridging to ", cfg.Address)
				conn, err := proto.Dial(cfg)
				if err == nil {
					err = app.Broker.ListenOnBridge(conn)
				}
				app.Log.Errorln("[server] bridge error:", err)
				time.Sleep(5 * time.Second)
			}
		}(cfg)
	}

	// Enable each module listed in the config
	for _, module := range cfg.EnabledModules {
		app.Must(app.EnableModule(module))
	}

	app.WriteConfig()
	core.WaitUntilInterrupt()
}
