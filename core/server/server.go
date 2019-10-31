// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package server provides a broker and modular host for microservices.
package server

import (
	"fmt"
	"os"
	"time"

	"github.com/sarifsystems/sarif/core"
	"github.com/sarifsystems/sarif/pkg/inject"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
	config "github.com/sarifsystems/sarif/services/schema"
	"github.com/sarifsystems/sarif/transports/sfproto"
)

type Server struct {
	*core.App
	ServerConfig Config

	Broker *sfproto.Broker
	Client sarif.Client
	*services.ModuleManager
	configStoreInitialized bool
}

type Config struct {
	Name           string
	ConfigStore    string
	Listen         []*sfproto.NetConfig
	Bridges        []*sfproto.NetConfig
	Gateways       []*sfproto.NetConfig
	EnabledModules []string
	BaseModules    []string
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

func (s *Server) InitBroker() error {
	if s.Broker != nil {
		return nil
	}

	sfproto.SetDefaultLogger(s.Log)
	s.Broker = sfproto.NewBroker()
	s.ClientFactory = s.Broker
	if s.Log.GetLevel() <= core.LogLevelTrace {
		s.Broker.TraceMessages(true)
	}

	client, err := s.ClientFactory.NewClient(sarif.ClientInfo{
		Name: s.ServerConfig.Name + "/sarifd",
	})
	if err != nil {
		return err
	}
	s.Client = client

	cfg := &s.ServerConfig
	if _, ok := s.Config.Get("server", cfg); !ok {
		if len(cfg.Listen) == 0 {
			cfg.Listen = append(cfg.Listen, &sfproto.NetConfig{
				Address: "tcp://localhost:23100",
				Auth:    sfproto.AuthNone,
			})
			s.Config.Set("server", cfg)
		}
	}

	// Listen on connections
	for _, cfg := range cfg.Listen {
		go func(cfg *sfproto.NetConfig) {
			s.Log.Infoln("[server] listening on", cfg.Address)
			s.Must(s.Broker.Listen(cfg))
		}(cfg)
	}

	// Setup bridges
	for _, cfg := range cfg.Bridges {
		go func(cfg *sfproto.NetConfig) {
			for {
				s.Log.Infoln("[server] bridging to ", cfg.Address)
				conn, err := sfproto.RawDial(cfg)
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
		go func(cfg *sfproto.NetConfig) {
			for {
				s.Log.Infoln("[server] gateway to ", cfg.Address)
				conn, err := sfproto.RawDial(cfg)
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
	cname := name
	if s.ServerConfig.Name != "" {
		cname = s.ServerConfig.Name + "/" + name
	}

	c, err := s.Broker.NewClient(sarif.ClientInfo{
		Name: cname,
	})
	s.Must(err)

	inj.Instance(s.Broker)
	inj.Factory(func() sarif.ClientFactory {
		return s.Broker
	})
	inj.Factory(func() sarif.Client {
		return c
	})
	inj.Factory(func() services.Config {
		if !s.configStoreInitialized {
			return s.Config.Section(name)
		}
		cfg := config.NewConfigStore(c)
		cfg.Store.StoreName = s.ServerConfig.ConfigStore
		cfg.ConfigDir = s.Config.Dir()
		return cfg
	})
}

func (s *Server) Inject(name string, container interface{}) error {
	inj := inject.NewInjector()
	s.SetupInjector(inj, name)
	return inj.Inject(container)
}

func (s *Server) InitModules() error {
	// We need to initialize some modules before others
	// TODO: Real dependency support
	for _, module := range s.ServerConfig.BaseModules {
		if err := s.EnableModule(module); err != nil {
			return err
		}
	}

	if err := s.findConfigStore(); err != nil {
		return err
	}

	for _, module := range s.ServerConfig.EnabledModules {
		if err := s.EnableModule(module); err != nil {
			s.Log.Infoln(module)
			return err
		}
	}
	return nil
}

func (s *Server) findConfigStore() error {
	time.Sleep(100 * time.Millisecond)
	for try := 1; try <= 5; try++ {
		req := s.Client.Request(sarif.Message{
			Action:      "proto/discover/store/get/config",
			Destination: s.ServerConfig.ConfigStore,
		})
		select {
		case msg := <-req:
			if s.ServerConfig.ConfigStore != msg.Source {
				s.ServerConfig.ConfigStore = msg.Source
				s.Config.Set("server", &s.ServerConfig)
			}
			s.configStoreInitialized = true
			return nil
		case <-time.After(1 * time.Duration(try) * time.Second):
		}
	}

	if s.ServerConfig.ConfigStore != "" {
		return fmt.Errorf("Could not connect to store %q", s.ServerConfig.ConfigStore)
	}
	return nil
}

func (s *Server) instantiate(m *services.Module) (interface{}, error) {
	inj := inject.NewInjector()
	s.SetupInjector(inj, m.Name)
	return inj.Create(m.NewInstance)
}
