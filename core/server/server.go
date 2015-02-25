// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

type Server struct {
	*core.App
	Broker    *proto.Broker
	Database  *core.DB
	Orm       *core.Orm
	Modules   map[string]*services.Module
	Instances map[string]interface{}
}

func New(appName, moduleName string) *Server {
	if moduleName == "" {
		moduleName = "server"
	}
	app := core.NewApp(appName, moduleName)
	return &Server{
		App:       app,
		Modules:   make(map[string]*services.Module),
		Instances: make(map[string]interface{}),
	}
}

func Init(appName, moduleName string) *Server {
	s := New(appName, moduleName)
	s.Init()
	return s
}

func (s *Server) InitDatabase() error {
	if s.Orm != nil {
		return nil
	}

	cfg := core.DatabaseConfig{
		Driver: "sqlite3",
		Source: core.GetDefaultDir(s.AppName) + "/" + s.ModuleName + ".db",
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
	s.Database = db.Database()
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
	return nil
}

func (s *Server) SetupInjector(inj *core.Injector, name string) {
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
			c := proto.NewClient(name)
			c.Connect(s.Broker.NewLocalConn())
			c.SetLogger(s.Log)
			return c
		})
	}
}

func (s *Server) Inject(name string, container interface{}) error {
	inj := core.NewInjector()
	s.SetupInjector(inj, name)
	return inj.Inject(container)
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

	inj := core.NewInjector()
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
		return i.Disable()
	}
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
