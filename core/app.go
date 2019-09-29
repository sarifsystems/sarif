// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package core handles most of the common db, config and proto initialization.
package core

import (
	"flag"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/sarifmq"
	"github.com/sarifsystems/sarif/sfproto"
)

var verbose = flag.Bool("v", false, "verbose debug output")
var vverbose = flag.Bool("vv", false, "very verbose debug output: db, individual messages")
var configPath = flag.String("config", "", "path to config file")

type App struct {
	AppName    string
	ModuleName string
	Config     *Config
	Log        *Logger
}

func NewApp(appName, moduleName string) *App {
	if appName == "" {
		appName = "sarif"
	}
	if !flag.Parsed() {
		flag.Parse()
	}

	app := &App{
		AppName:    appName,
		ModuleName: moduleName,
		Log:        DefaultLog,
	}
	if *vverbose {
		app.Log.SetFlags(log.Ldate | log.Lmicroseconds)
		app.Log.SetLevel(LogLevelTrace)
	} else if *verbose {
		app.Log.SetLevel(LogLevelDebug)
	} else {
		app.Log.SetLevel(LogLevelInfo)
	}

	return app
}

func (app *App) Init() {
	if err := app.initConfig(); err != nil {
		app.Log.Fatalln(err)
	}

	app.WriteConfig()
}

func (app *App) initConfig() (err error) {
	if *configPath == "" {
		if app.ModuleName == "temp" {
			dir, err := ioutil.TempDir(os.TempDir(), app.AppName+"-")
			if err != nil {
				return err
			}
			app.Config = NewConfig(dir + "/config.json")
		} else {
			app.Config, err = FindConfig(app.AppName, app.ModuleName)
		}
	} else {
		app.Config, err = OpenConfig(*configPath, true)
	}
	if err != nil {
		return err
	}
	app.Log.Debugf("[core] reading config from '%s'", app.Config.Path())
	return nil
}

func (app *App) WriteConfig() {
	if app.Config.IsModified() {
		app.Log.Infof("[core] writing config to '%s'", app.Config.Path())
		app.Must(app.Config.Write())
	}
}

func (app *App) Close() {
	if app.ModuleName == "temp" {
		app.Must(os.RemoveAll(app.Config.Dir()))
	} else {
		app.WriteConfig()
	}
}

func (app *App) Must(err error) {
	if err != nil {
		app.Log.Fatalln(err)
	}
}

func (app *App) ClientDial(ci sarif.ClientInfo) (sarif.Client, error) {
	cfg := sfproto.NetConfig{
		Address: "tcp://localhost:" + sfproto.DefaultPort,
	}
	app.Config.Get("dial", &cfg)
	app.WriteConfig()

	u, err := url.Parse(cfg.Address)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "amqp" {
		return app.ClientDialAmqp(ci, cfg)
	}

	return app.ClientDialProto(ci, cfg)
}

func (app *App) ClientDialProto(ci sarif.ClientInfo, cfg sfproto.NetConfig) (*sfproto.Client, error) {
	c := sfproto.NewClient(ci.Name)
	c.Info = ci
	c.OnConnectionLost(func(err error) {
		app.Log.Errorln("connection lost:", err)
		for {
			time.Sleep(5 * time.Second)
			if err := c.Dial(&cfg); err == nil {
				app.Log.Println("reconnected")
				return
			}
		}
	})

	err := c.Dial(&cfg)
	return c, err
}

func (app *App) ClientDialAmqp(ci sarif.ClientInfo, cfg sfproto.NetConfig) (*sarifmq.Client, error) {
	c := sarifmq.NewClient(ci.Name)
	c.Info = ci
	err := c.Dial(&cfg)
	return c, err
}

func WaitUntilInterrupt() {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-ch
}
