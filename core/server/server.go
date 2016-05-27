// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package server provides a broker and modular host for microservices.
package server

import (
	"os"
	"time"

	"github.com/sarifsystems/sarif/core"
	"github.com/sarifsystems/sarif/pkg/inject"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
)

type Server struct {
	*core.App
	ServerConfig Config

	Broker *sarif.Broker
	Orm    *core.Orm
	*services.ModuleManager
}

type Config struct {
	Name           string
	Listen         []*sarif.NetConfig
	Bridges        []*sarif.NetConfig
	Gateways       []*sarif.NetConfig
	EnabledModules []string
}

func New(appName, moduleName string) *Server {
	if moduleName == "" {
		moduleName = "server"
	}
	app := core.NewApp(appName, moduleName)
	s := &Server{
		App: app,
	}
	s.ModuleManager = services.NewModuleManager(s.instantiate)
	if n, err := os.Hostname(); err == nil {
		s.ServerConfig.Name = n
	}
	return s
}

func (s *Server) Init() {
	s.App.Init()
	s.Must(s.InitDatabase())
	s.Must(s.InitBroker())
	s.Must(s.InitModules())
	s.WriteConfig()
}

func (s *Server) Close() {
	for _, module := range s.ServerConfig.EnabledModules {
		s.DisableModule(module)
	}

	s.App.Close()
}

func (s *Server) Run() {
	s.Init()
	core.WaitUntilInterrupt()
	defer s.Close()
}

func (s *Server) InitDatabase() error {
	if s.Orm != nil {
		return nil
	}

	cfg := core.DatabaseConfig{
		Driver: "sqlite3",
		Source: s.Config.Dir() + "/" + s.ModuleName + ".db",
	}

	s.Config.Get("database", &cfg)

	db, err := core.OpenDatabase(cfg)
	if err != nil {
		return err
	}
	if s.Log.GetLevel() <= core.LogLevelTrace {
		db.LogMode(true)
	}
	s.Orm = db
	return nil
}

func (s *Server) InitBroker() error {
	if s.Broker != nil {
		return nil
	}

	sarif.SetDefaultLogger(s.Log)
	s.Broker = sarif.NewBroker()
	if s.Log.GetLevel() <= core.LogLevelTrace {
		s.Broker.TraceMessages(true)
	}

	cfg := &s.ServerConfig
	if _, ok := s.Config.Get("server", cfg); !ok {
		if len(cfg.Listen) == 0 {
			cfg.Listen = append(cfg.Listen, &sarif.NetConfig{
				Address: "tcp://localhost:23100",
			})
			s.Config.Set("server", cfg)
		}
	}

	// Listen on connections
	for _, cfg := range cfg.Listen {
		go func(cfg *sarif.NetConfig) {
			s.Log.Infoln("[server] listening on", cfg.Address)
			s.Must(s.Broker.Listen(cfg))
		}(cfg)
	}

	// Setup bridges
	for _, cfg := range cfg.Bridges {
		go func(cfg *sarif.NetConfig) {
			for {
				s.Log.Infoln("[server] bridging to ", cfg.Address)
				conn, err := sarif.Dial(cfg)
				if err == nil {
					err = s.Broker.ListenOnBridge(conn)
				}
				s.Log.Errorln("[server] bridge error:", err)
				time.Sleep(5 * time.Second)
			}
		}(cfg)
	}

	// Setup gateways
	for _, cfg := range cfg.Gateways {
		go func(cfg *sarif.NetConfig) {
			for {
				s.Log.Infoln("[server] gateway to ", cfg.Address)
				conn, err := sarif.Dial(cfg)
				if err == nil {
					err = s.Broker.ListenOnGateway(conn)
				}
				s.Log.Errorln("[server] gateway error:", err)
				time.Sleep(5 * time.Second)
			}
		}(cfg)
	}

	return nil
}

func (s *Server) SetupInjector(inj *inject.Injector, name string) {
	s.App.SetupInjector(inj, name)
	if s.Orm != nil {
		inj.Instance(s.Orm.DB)
		inj.Instance(s.Orm.Database())
	}
	if s.Broker != nil {
		inj.Instance(s.Broker)
		inj.Factory(func() sarif.Conn {
			return s.Broker.NewLocalConn()
		})
		inj.Factory(func() *sarif.Client {
			cname := name
			if s.ServerConfig.Name != "" {
				cname = s.ServerConfig.Name + "/" + name
			}
			c := sarif.NewClient(cname)
			c.Connect(s.Broker.NewLocalConn())
			c.SetLogger(s.Log)
			return c
		})
	}
}

func (s *Server) Inject(name string, container interface{}) error {
	inj := inject.NewInjector()
	s.SetupInjector(inj, name)
	return inj.Inject(container)
}

func (s *Server) InitModules() error {
	for _, module := range s.ServerConfig.EnabledModules {
		if err := s.EnableModule(module); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) instantiate(m *services.Module) (interface{}, error) {
	inj := inject.NewInjector()
	s.SetupInjector(inj, m.Name)
	return inj.Create(m.NewInstance)
}
