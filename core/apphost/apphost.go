// Copyright (C) 2019 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package apphost provides a modular host for microservices.
package apphost

import (
	"fmt"
	"os"
	"time"

	"github.com/sarifsystems/sarif/core"
	"github.com/sarifsystems/sarif/pkg/inject"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
	config "github.com/sarifsystems/sarif/services/schema"
)

type AppHost struct {
	*core.App
	HostConfig Config
	Client     sarif.Client
	*services.ModuleManager
	configStoreInitialized bool
}

type Config struct {
	Name           string
	ConfigStore    string
	EnabledModules []string
	BaseModules    []string
}

func New(appName, moduleName string) *AppHost {
	if moduleName == "" {
		moduleName = "apphost"
	}
	app := core.NewApp(appName, moduleName)
	s := &AppHost{
		App: app,
	}
	s.ModuleManager = services.NewModuleManager(s.instantiate)
	if n, err := os.Hostname(); err == nil {
		s.HostConfig.Name = n
	}
	return s
}

func (s *AppHost) Init() {
	s.App.Init()
	s.InitClient()
	s.Must(s.InitModules())
	s.WriteConfig()
}

func (s *AppHost) InitClient() {
	s.InitClientFactory()
	c, err := s.ClientFactory.NewClient(sarif.ClientInfo{
		Name: s.HostConfig.Name + "/sarifd",
	})
	s.Must(err)
	s.Client = c
}

func (s *AppHost) Close() {
	for _, module := range s.HostConfig.EnabledModules {
		s.DisableModule(module)
	}

	s.App.Close()
}

func (s *AppHost) Run() {
	s.Init()
	core.WaitUntilInterrupt()
	defer s.Close()
}

func (s *AppHost) SetupInjector(inj *inject.Injector, name string) {
	cname := name
	if s.HostConfig.Name != "" {
		cname = s.HostConfig.Name + "/" + name
	}

	c, err := s.ClientFactory.NewClient(sarif.ClientInfo{
		Name: cname,
	})
	s.Must(err)

	inj.Factory(func() sarif.ClientFactory {
		return s.ClientFactory
	})
	inj.Factory(func() sarif.Client {
		return c
	})
	inj.Factory(func() services.Config {
		if !s.configStoreInitialized {
			return s.Config.Section(name)
		}
		cfg := config.NewConfigStore(c)
		cfg.Store.StoreName = s.HostConfig.ConfigStore
		cfg.ConfigDir = s.Config.Dir()
		return cfg
	})
}

func (s *AppHost) Inject(name string, container interface{}) error {
	inj := inject.NewInjector()
	s.SetupInjector(inj, name)
	return inj.Inject(container)
}

func (s *AppHost) InitModules() error {
	// We need to initialize some modules before others
	// TODO: Real dependency support
	for _, module := range s.HostConfig.BaseModules {
		if err := s.EnableModule(module); err != nil {
			return err
		}
	}

	if err := s.findConfigStore(); err != nil {
		return err
	}

	for _, module := range s.HostConfig.EnabledModules {
		fmt.Println(module)
		if err := s.EnableModule(module); err != nil {
			s.Log.Infoln(module)
			return err
		}
		fmt.Println("done")
	}
	return nil
}

func (s *AppHost) findConfigStore() error {
	time.Sleep(100 * time.Millisecond)
	for try := 1; try <= 5; try++ {
		req := s.Client.Request(sarif.Message{
			Action:      "proto/discover/store/get/config",
			Destination: s.HostConfig.ConfigStore,
		})
		select {
		case msg := <-req:
			if s.HostConfig.ConfigStore != msg.Source {
				s.HostConfig.ConfigStore = msg.Source
				s.Config.Set("server", &s.HostConfig)
			}
			s.configStoreInitialized = true
			return nil
		case <-time.After(1 * time.Duration(try) * time.Second):
		}
	}

	if s.HostConfig.ConfigStore != "" {
		return fmt.Errorf("Could not connect to store %q", s.HostConfig.ConfigStore)
	}
	return nil
}

func (s *AppHost) instantiate(m *services.Module) (interface{}, error) {
	inj := inject.NewInjector()
	s.SetupInjector(inj, m.Name)
	return inj.Create(m.NewInstance)
}
