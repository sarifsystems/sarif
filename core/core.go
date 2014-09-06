// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"os"
	"os/signal"

	"github.com/xconstruct/stark/conf"
	"github.com/xconstruct/stark/database"
	"github.com/xconstruct/stark/log"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/proto/transports/mqtt"
)

type App struct {
	AppName  string
	Config   *conf.Config
	Proto    *proto.Mux
	Database *database.DB
	Log      *log.Logger

	instances map[string]ModuleInstance
}

func NewApp(appName string) (*App, error) {
	app := &App{
		AppName:   appName,
		Log:       log.Default,
		instances: make(map[string]ModuleInstance),
	}
	app.Log.SetLevel(log.LevelInfo)

	if err := app.initConfig(); err != nil {
		return app, err
	}
	if err := app.initDatabase(); err != nil {
		return app, err
	}
	if err := app.initProto(); err != nil {
		return app, err
	}

	app.writeConfig()

	return app, nil
}

func (app *App) GetDefaultDir() string {
	path := os.Getenv("XDG_CONFIG_HOME")
	if path != "" {
		return path + "/" + app.AppName
	}

	home := os.Getenv("HOME")
	if home != "" {
		return home + "/.config/" + app.AppName
	}

	return "."
}

func (app *App) initConfig() error {
	f := app.GetDefaultDir() + "/config.json"
	app.Log.Debugf("[core] reading config from '%s'", f)
	cfg, err := conf.Read(f)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		cfg = conf.New()
		app.Log.Warnf("[core] config not found, loading defaults")
		if err := conf.Write(f, cfg); err != nil {
			return err
		}

	}
	app.Config = cfg
	return nil
}

func (app *App) writeConfig() {
	if app.Config.IsModified() {
		f := app.GetDefaultDir() + "/config.json"
		app.Log.Infof("[core] writing config to '%s'", f)
		app.Must(conf.Write(f, app.Config))
	}
}

func (app *App) Close() {
	app.writeConfig()
}

func (app *App) initDatabase() error {
	cfg := database.Config{
		Driver: "sqlite3",
		Source: app.GetDefaultDir() + "/" + app.AppName + ".db",
	}

	if err := app.Config.Get("database", &cfg); err != nil {
		return err
	}

	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	app.Database = db
	return nil
}

func (app *App) initProto() error {
	cfg := mqtt.GetDefaults()
	if err := app.Config.Get("mqtt", &cfg); err != nil {
		return err
	}
	app.Proto = proto.NewMux()

	if cfg.Server == "" || cfg.Server == "tcp://example.org:1883" {
		app.Log.Warnln("[core] config 'mqtt.Server' empty, falling back to local broker")
		app.Proto.RegisterPublisher(func(msg proto.Message) error {
			raw, _ := msg.Encode()
			app.Log.Debugln("[core] broker received:", string(raw))
			app.Proto.Handle(msg)
			return nil
		})
		return nil
	}

	m := mqtt.New(cfg)
	proto.Connect(m, app.Proto)
	return m.Connect()
}

func (app *App) Must(err error) {
	if err != nil {
		app.Log.Fatalln(err)
	}
}

func (app *App) EnableModule(name string) error {
	i, ok := app.instances[name]
	if ok {
		return nil
	}

	m, err := GetModule(name)
	if err != nil {
		return err
	}
	ctx := app.NewContext()
	i, err = m.NewInstance(ctx)
	app.instances[name] = i
	if err != nil {
		return err
	}
	app.Log.Infof("[core] module '%s' enabled", name)
	return i.Enable()
}

func (app *App) DisableModule(name string) error {
	i, ok := app.instances[name]
	if !ok {
		return nil
	}
	app.instances[name] = nil
	app.Log.Infof("[core] module '%s' disabled", name)
	return i.Disable()
}

func (app *App) WaitUntilInterrupt() {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt)
	<-ch
}

func (app *App) NewContext() *Context {
	return &Context{
		app.Database,
		app.Log,
		app.Proto.NewEndpoint(),
		app.Config,
	}
}
