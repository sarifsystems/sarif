// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"os"
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/pkg/inject"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

type Server struct {
	*core.App
	ServerConfig Config

	Broker    *proto.Broker
	Orm       *core.Orm
	Modules   map[string]*services.Module
	Instances map[string]interface{}
}

type Config struct {
	Name           string
	Listen         []*proto.NetConfig
	Bridges        []*proto.NetConfig
	Gateways       []*proto.NetConfig
	EnabledModules []string
}

func New(appName, moduleName string) *Server {
	if moduleName == "" {
		moduleName = "server"
	}
	app := core.NewApp(appName, moduleName)
	s := &Server{
		App:       app,
		Modules:   make(map[string]*services.Module),
		Instances: make(map[string]interface{}),
	}
	if n, err := os.Hostname(); err != nil {
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

	proto.SetDefaultLogger(s.Log)
	s.Broker = proto.NewBroker()
	if s.Log.GetLevel() <= core.LogLevelTrace {
		s.Broker.TraceMessages(true)
	}

	cfg := &s.ServerConfig
	if _, ok := s.Config.Get("server", cfg); !ok {
		if len(cfg.Listen) == 0 {
			cfg.Listen = append(cfg.Listen, &proto.NetConfig{
				Address: "tcp://localhost:23100",
			})
			s.Config.Set("server", cfg)
		}
	}

	// Listen on connections
	for _, cfg := range cfg.Listen {
		go func(cfg *proto.NetConfig) {
			s.Log.Infoln("[server] listening on", cfg.Address)
			s.Must(s.Broker.Listen(cfg))
		}(cfg)
	}

	// Setup bridges
	for _, cfg := range cfg.Bridges {
		go func(cfg *proto.NetConfig) {
			for {
				s.Log.Infoln("[server] bridging to ", cfg.Address)
				conn, err := proto.Dial(cfg)
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
		go func(cfg *proto.NetConfig) {
			for {
				s.Log.Infoln("[server] gateway to ", cfg.Address)
				conn, err := proto.Dial(cfg)
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
		inj.Factory(func() proto.Conn {
			return s.Broker.NewLocalConn()
		})
		inj.Factory(func() *proto.Client {
			cname := name
			if s.ServerConfig.Name != "" {
				cname = s.ServerConfig.Name + "/" + name
			}
			c := proto.NewClient(cname)
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

func (s *Server) EnableModule(name string) error {
	i, ok := s.Instances[name]
	if ok {
		return nil
	}

	m, err := s.GetModule(name)
	if err != nil {
		return err
	}

	inj := inject.NewInjector()
	s.SetupInjector(inj, name)
	i, err = inj.Create(m.NewInstance)
	if err != nil {
		return err
	}
	s.Instances[name] = i
	s.Log.Infof("[core] module '%s' enabled", name)

	if i, ok := i.(enabler); ok {
		return i.Enable()
	}
	return nil
}

func (s *Server) DisableModule(name string) error {
	i, ok := s.Instances[name]
	if !ok {
		return nil
	}
	s.Instances[name] = nil
	s.Log.Infof("[core] module '%s' disabled", name)
	if i, ok := i.(disabler); ok {
		if err := i.Disable(); err != nil {
			return err
		}
	}
	delete(s.Instances, name)
	return nil
}

type ErrModuleNotFound struct {
	Module string
}

func (e ErrModuleNotFound) Error() string {
	return "module '" + e.Module + "' not found'"
}

func (s *Server) RegisterModule(mod *services.Module) {
	s.Modules[mod.Name] = mod
}

func (s *Server) GetModule(name string) (*services.Module, error) {
	m, ok := s.Modules[name]
	if !ok {
		return m, ErrModuleNotFound{name}
	}
	return m, nil
}

type enabler interface {
	Enable() error
}

type disabler interface {
	Disable() error
}
