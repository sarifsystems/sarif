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
	"github.com/xconstruct/stark/services"
)

var verbose = flag.Bool("v", false, "verbose debug output")
var vverbose = flag.Bool("vv", false, "very verbose debug output: db, individual messages")
var configPath = flag.String("config", "", "path to config file")

type App struct {
	AppName  string
	Config   *Config
	Broker   *proto.Broker
	Database *DB
	Orm      *Orm
	Log      *Logger

	modules   map[string]*services.Module
	instances map[string]interface{}
}

func NewApp(appName string) *App {
	if !flag.Parsed() {
		flag.Parse()
	}

	app := &App{
		AppName:   appName,
		Log:       DefaultLog,
		modules:   make(map[string]*services.Module),
		instances: make(map[string]interface{}),
	}
	app.Log.SetLevel(LogLevelInfo)
	if *verbose || *vverbose {
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
	if err := app.initBroker(); err != nil {
		return err
	}

	app.writeConfig()
	return nil
}

func (app *App) initConfig() error {
	path := *configPath
	if path == "" {
		path = GetDefaultDir(app.AppName) + "/config.json"
	}
	cfg, err := OpenConfig(path, true)
	if err != nil {
		return err
	}
	app.Log.Debugf("[core] reading config from '%s'", cfg.Path())
	app.Config = cfg
	return nil
}

func (app *App) writeConfig() {
	if app.Config.IsModified() {
		app.Log.Infof("[core] writing config to '%s'", app.Config.Path())
		app.Must(app.Config.Write())
	}
}

func (app *App) Close() {
	app.writeConfig()
}

func (app *App) initDatabase() error {
	cfg := DatabaseConfig{
		Driver: "sqlite3",
		Source: GetDefaultDir(app.AppName) + "/" + app.AppName + ".db",
	}

	app.Config.Get("database", &cfg)

	db, err := OpenDatabase(cfg)
	if err != nil {
		return err
	}
	if *vverbose {
		db.LogMode(true)
	}
	app.Orm = db
	app.Database = db.Database()
	return nil
}

func (app *App) initBroker() error {
	proto.SetDefaultLogger(app.Log)
	app.Broker = proto.NewBroker()
	if *vverbose {
		app.Broker.TraceMessages(true)
	}
	return nil
}

func (app *App) Must(err error) {
	if err != nil {
		app.Log.Fatalln(err)
	}
}

func (app *App) setupInjector(name string) *Injector {
	inj := NewInjector()
	inj.Instance(app.Config)
	inj.Instance(app.Orm.DB)
	inj.Instance(app.Orm.Database())
	inj.Instance(app.Log)
	inj.Instance(app.Broker)
	inj.Factory(func() proto.Logger {
		return app.Log
	})
	inj.Factory(func() proto.Conn {
		return app.Broker.NewLocalConn()
	})
	inj.Factory(func() *proto.Client {
		conn := app.Broker.NewLocalConn()
		c := proto.NewClient(name, conn)
		c.SetLogger(app.Log)
		return c
	})
	return inj
}

func (app *App) Inject(name string, container interface{}) error {
	inj := app.setupInjector(name)
	return inj.Inject(container)
}

func (app *App) EnableModule(name string) error {
	i, ok := app.instances[name]
	if ok {
		return nil
	}

	m, err := app.GetModule(name)
	if err != nil {
		return err
	}

	inj := app.setupInjector(name)
	i, err = inj.Create(m.NewInstance)
	app.instances[name] = i
	if err != nil {
		return err
	}
	app.Log.Infof("[core] module '%s' enabled", name)

	if i, ok := i.(enabler); ok {
		return i.Enable()
	}
	return nil
}

func (app *App) DisableModule(name string) error {
	i, ok := app.instances[name]
	if !ok {
		return nil
	}
	app.instances[name] = nil
	app.Log.Infof("[core] module '%s' disabled", name)
	if i, ok := i.(disabler); ok {
		return i.Disable()
	}
	return nil
}

func WaitUntilInterrupt() {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt)
	<-ch
}

func (app *App) NewContext() *Context {
	return &Context{
		app.Database,
		app.Orm,
		app.Log,
		app.Broker.NewLocalConn(),
		app.Config,
	}
}
