// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"flag"
	"os"
	"os/signal"

	"github.com/xconstruct/stark/proto"
)

var verbose = flag.Bool("v", false, "verbose debug output")

type App struct {
	AppName  string
	Config   *Config
	Proto    *proto.Mux
	Database *DB
	Orm      *Orm
	Log      *Logger

	instances map[string]ModuleInstance
}

func NewApp(appName string) *App {
	if !flag.Parsed() {
		flag.Parse()
	}

	app := &App{
		AppName:   appName,
		Log:       DefaultLog,
		instances: make(map[string]ModuleInstance),
	}
	app.Log.SetLevel(LogLevelInfo)
	if *verbose {
		app.Log.SetLevel(LogLevelDebug)
	}

	return app
}

func (app *App) Init() error {
	if err := app.initConfig(); err != nil {
		return err
	}
	if err := app.initDatabase(); err != nil {
		return err
	}
	if err := app.initProto(); err != nil {
		return err
	}

	app.writeConfig()
	return nil
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
	cfg, err := ReadConfig(f)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		cfg = NewConfig()
		app.Log.Warnf("[core] config not found, loading defaults")
		if err := WriteConfig(f, cfg); err != nil {
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
		app.Must(WriteConfig(f, app.Config))
	}
}

func (app *App) Close() {
	app.writeConfig()
}

func (app *App) initDatabase() error {
	cfg := DatabaseConfig{
		Driver: "sqlite3",
		Source: app.GetDefaultDir() + "/" + app.AppName + ".db",
	}

	if err := app.Config.Get("database", &cfg); err != nil {
		return err
	}

	db, err := OpenDatabase(cfg)
	if err != nil {
		return err
	}
	app.Orm = db
	app.Database = db.Database()
	return nil
}

func (app *App) initProto() error {
	proto.SetDefaultLogger(app.Log)
	cfg := proto.GetMqttDefaults()
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

	app.Log.Debugf("[core] mqtt connecting to %s", cfg.Server)
	m := proto.DialMqtt(cfg)
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
		app.Orm,
		app.Log,
		app.Proto.NewConn(),
		app.Config,
	}
}
