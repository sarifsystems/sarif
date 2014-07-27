package core

import (
	"os"
	"os/signal"

	"github.com/xconstruct/stark/conf"
	"github.com/xconstruct/stark/database"
	"github.com/xconstruct/stark/log"
	"github.com/xconstruct/stark/proto/client"
	"github.com/xconstruct/stark/proto/transports/mqtt"
)

type Context struct {
	AppName  string
	Config   *conf.Config
	Proto    *client.Client
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
		c.Log.Debugf("[core] writing config to '%s'", f)
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

	c.Proto = client.New(c.AppName)
	c.Proto.SetTransport(mqtt.New(cfg))
	return nil
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
