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
	"github.com/xconstruct/stark/proto/mux"
	"github.com/xconstruct/stark/proto/transports/mqtt"
)

type Context struct {
	AppName  string
	Config   *conf.Config
	Proto    *mux.TransportMux
	Database *database.DB
	Log      *log.Logger

	instances map[string]ModuleInstance
}

func NewContext(appName string) (*Context, error) {
	c := &Context{
		AppName:   appName,
		Log:       log.Default,
		instances: make(map[string]ModuleInstance),
	}
	c.Log.SetLevel(log.LevelInfo)

	if err := c.initConfig(); err != nil {
		return c, err
	}
	if err := c.initDatabase(); err != nil {
		return c, err
	}
	if err := c.initProto(); err != nil {
		return c, err
	}

	return c, nil
}

func (c *Context) Close() {
	if c.Config.IsModified() {
		f := c.GetDefaultDir() + "/config.json"
		c.Log.Infof("[core] writing config to '%s'", f)
		c.Must(conf.Write(f, c.Config))
	}
}

func (c *Context) GetDefaultDir() string {
	path := os.Getenv("XDG_CONFIG_HOME")
	if path != "" {
		return path + "/" + c.AppName
	}

	home := os.Getenv("HOME")
	if home != "" {
		return home + "/.config/" + c.AppName
	}

	return "."
}

func (c *Context) initConfig() error {
	f := c.GetDefaultDir() + "/config.json"
	c.Log.Debugf("[core] reading config from '%s'", f)
	cfg, err := conf.Read(f)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		cfg = conf.New()
		c.Log.Warnf("[core] config not found, writing defaults to '%s'", f)
		if err := conf.Write(f, cfg); err != nil {
			return err
		}

	}
	c.Config = cfg
	return nil
}

func (c *Context) initDatabase() error {
	cfg := database.Config{
		Driver: "sqlite3",
		Source: c.GetDefaultDir() + "/" + c.AppName + ".db",
	}

	if err := c.Config.Get("database", &cfg); err != nil {
		return err
	}

	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	c.Database = db
	return nil
}

func (c *Context) initProto() error {
	cfg := mqtt.Config{}
	if err := c.Config.Get("mqtt", &cfg); err != nil {
		return err
	}
	c.Proto = mux.NewTransportMux()

	if cfg.Server == "" {
		c.Log.Warnln("[core] config 'mqtt.Server' empty, falling back to local broker")
		c.Proto.RegisterPublisher(func(msg proto.Message) error {
			c.Proto.Handle(msg)
			return nil
		})
		return nil
	}

	m := mqtt.New(cfg)
	proto.Connect(m, c.Proto)
	return m.Connect()
}

func (c *Context) NewProtoClient(deviceName string) *proto.Client {
	return proto.NewClient(deviceName, c.Proto.NewEndpoint())
}

func (c *Context) Must(err error) {
	if err != nil {
		c.Log.Fatalln(err)
	}
}

func (c *Context) EnableModule(name string) error {
	i, ok := c.instances[name]
	if ok {
		return nil
	}

	m, err := GetModule(name)
	if err != nil {
		return err
	}
	i, err = m.NewInstance(c)
	c.instances[name] = i
	if err != nil {
		return err
	}
	c.Log.Infof("[core] module '%s' enabled", name)
	return i.Enable()
}

func (c *Context) DisableModule(name string) error {
	i, ok := c.instances[name]
	if !ok {
		return nil
	}
	c.instances[name] = nil
	c.Log.Infof("[core] module '%s' disabled", name)
	return i.Disable()
}

func (c *Context) WaitUntilInterrupt() {
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, os.Interrupt)
	<-ch
}
