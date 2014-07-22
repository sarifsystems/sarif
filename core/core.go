package core

import (
	"github.com/xconstruct/stark/conf"
	"github.com/xconstruct/stark/database"
	"github.com/xconstruct/stark/log"
	"github.com/xconstruct/stark/proto/client"
	"github.com/xconstruct/stark/proto/transports/mqtt"
)

type Context struct {
	cfg    *conf.Config
	client *client.Client
	db     *database.DB
	Log    *log.Logger
}

func NewContext() (*Context, error) {
	return &Context{
		Log: log.Default,
	}, nil
}

func (c *Context) Config() (*conf.Config, error) {
	if c.cfg != nil {
		return c.cfg, nil
	}

	cfg, err := conf.ReadDefault()
	if err != nil {
		return nil, err
	}
	c.cfg = &cfg
	return c.cfg, nil
}

func (c *Context) Database() (*database.DB, error) {
	if c.db != nil {
		return c.db, nil
	}

	cfg, err := c.Config()
	if err != nil {
		return nil, err
	}
	db, err := database.Open(*cfg)
	if err != nil {
		return nil, err
	}
	c.db = db
	return c.db, nil
}

func (c *Context) Client() (*client.Client, error) {
	if c.client != nil {
		return c.client, nil
	}

	cfg, err := c.Config()
	if err != nil {
		return nil, err
	}
	client := client.New("server")
	client.SetTransport(mqtt.New(mqtt.Config{
		Server:      cfg.Proto.Server,
		Certificate: cfg.Proto.Certificate,
		Key:         cfg.Proto.Key,
		Authority:   cfg.Proto.Authority,
	}))
	if err := client.Connect(); err != nil {
		return nil, err
	}
	c.client = client
	return client, nil
}

func (c *Context) Must(err error) {
	if err != nil {
		c.Log.Fatalln(err)
	}
}
