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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sarifsystems/sarif/pkg/inject"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
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

func (app *App) SetupInjector(inj *inject.Injector, name string) {
	inj.Instance(app.Log)
	inj.Factory(func() services.Config {
		return app.Config.Section(name)
	})
	inj.Factory(func() sarif.Logger {
		return app.Log
	})
}

func (app *App) Inject(name string, container interface{}) error {
	inj := inject.NewInjector()
	app.SetupInjector(inj, name)
	return inj.Inject(container)
}

func (app *App) Dial() sarif.Conn {
	cfg := sarif.NetConfig{
		Address: "tcp://localhost:" + sarif.DefaultPort,
	}
	app.Config.Get("dial", &cfg)
	app.WriteConfig()
	conn, err := sarif.Dial(&cfg)
	if err != nil {
		app.Log.Fatal(err)
	}
	return conn
}

func (app *App) ClientDial(ci sarif.ClientInfo) (*sarif.Client, error) {
	cfg := sarif.NetConfig{
		Address: "tcp://localhost:" + sarif.DefaultPort,
	}
	app.Config.Get("dial", &cfg)
	app.WriteConfig()

	c := sarif.NewClient(ci.Name)
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

func WaitUntilInterrupt() {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-ch
}
