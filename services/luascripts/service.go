// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package luascripts

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "luascripts",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	ScriptDir string `json:"script_dir"`
}

type Dependencies struct {
	Config services.Config
	Log    proto.Logger
	Client *proto.Client
	Broker *proto.Broker
}

type Service struct {
	cfg    Config
	Log    proto.Logger
	Broker *proto.Broker
	*proto.Client

	Machines map[string]*Machine
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Log:      deps.Log,
		Broker:   deps.Broker,
		Machines: make(map[string]*Machine),
	}
	s.cfg.ScriptDir = deps.Config.Dir() + "/luascripts"
	deps.Config.Get(&s.cfg)

	sv := proto.NewClient("luascripts")
	sv.Connect(s.Broker.NewLocalConn())
	s.Client = sv

	s.createMachine("default")
	return s
}

func (s *Service) Enable() error {
	s.Subscribe("lua/do", "", s.handleLuaDo)
	s.Subscribe("cmd/lua", "", s.handleLuaDo)

	if s.cfg.ScriptDir == "" {
		return nil
	}
	dir, err := os.Open(s.cfg.ScriptDir)
	if err != nil {
		if os.IsNotExist(err) {
			s.Log.Warnln("[luascripts]:", err)
			return nil
		}
		return err
	}
	files, err := dir.Readdirnames(0)
	if err != nil {
		return err
	}
	for _, f := range files {
		if !strings.HasSuffix(f, ".lua") {
			continue
		}
		s.createMachineFromScript(f)
	}
	return nil
}

func (s *Service) createMachineFromScript(f string) (*Machine, error) {
	s.Log.Infoln("[luascripts] loading ", f)
	m, err := s.createMachine(strings.TrimSuffix(f, ".lua"))
	if err != nil {
		return m, err
	}
	err = m.Lua.DoFile(s.cfg.ScriptDir + "/" + f)
	return m, err
}

func (s *Service) createMachine(name string) (*Machine, error) {
	if name == "" {
		name = proto.GenerateId()
	}
	if _, ok := s.Machines[name]; ok {
		return nil, errors.New("Machine " + name + " already exists")
	}

	c := proto.NewClient("luascripts/" + name)
	c.RequestTimeout = 5 * time.Second
	c.Connect(s.Broker.NewLocalConn())

	m := NewMachine(s.Log, c)
	s.Machines[name] = m
	if err := m.Enable(); err != nil {
		return m, err
	}
	return m, nil
}

func (s *Service) destroyMachine(name string) error {
	m, ok := s.Machines[name]
	if !ok {
		return errors.New("Machine " + name + " does not exist")
	}
	delete(s.Machines, name)
	return m.Disable()
}

func (s *Service) handleLuaDo(msg proto.Message) {
	m := s.Machines["default"]
	out, err := m.Do(msg.Text)
	if err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	s.Reply(msg, proto.Message{
		Action: "lua/done",
		Text:   out,
	})
}
