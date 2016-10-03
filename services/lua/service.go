// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service lua provides Lua scripting for the sarif network.
package lua

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"

	"github.com/sarifsystems/sarif/pkg/content"
	"github.com/sarifsystems/sarif/pkg/schema"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
)

var Module = &services.Module{
	Name:        "lua",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	ScriptDir string `json:"script_dir"`
}

type Dependencies struct {
	Config services.Config
	Client *sarif.Client
	Broker *sarif.Broker
}

type Service struct {
	cfg    Config
	Broker *sarif.Broker
	*sarif.Client

	Machines map[string]*Machine
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Broker:   deps.Broker,
		Client:   deps.Client,
		Machines: make(map[string]*Machine),
	}
	s.cfg.ScriptDir = deps.Config.Dir() + "/lua"
	deps.Config.Get(&s.cfg)

	s.createMachine("default")
	return s
}

func (s *Service) Enable() error {
	s.Subscribe("lua/do", "", s.handleLuaDo)
	s.Subscribe("lua/start", "", s.handleLuaStart)
	s.Subscribe("lua/stop", "", s.handleLuaStop)
	s.Subscribe("lua/load", "", s.handleLuaLoad)
	s.Subscribe("lua/dump", "", s.handleLuaDump)
	s.Subscribe("cmd/lua", "", s.handleLuaDo)

	if s.cfg.ScriptDir == "" {
		return nil
	}
	dir, err := os.Open(s.cfg.ScriptDir)
	if err != nil {
		if os.IsNotExist(err) {
			s.Log("warn", err.Error())
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
	s.Log("info", "loading "+f)
	m, err := s.createMachine(strings.TrimSuffix(f, ".lua"))
	if err != nil {
		return m, err
	}
	err = m.Lua.DoFile(s.cfg.ScriptDir + "/" + f)
	return m, err
}

func (s *Service) createMachine(name string) (*Machine, error) {
	if name == "" {
		name = sarif.GenerateId()
	}
	if _, ok := s.Machines[name]; ok {
		return nil, errors.New("Machine " + name + " already exists")
	}

	c := sarif.NewClient(s.DeviceId + "/" + name)
	c.Connect(s.Broker.NewLocalConn())

	m := NewMachine(c)
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

func (s *Service) getOrCreateMachine(name string) (*Machine, error) {
	if m, ok := s.Machines[name]; ok {
		return m, nil
	}
	return s.createMachine(name)
}

func (s *Service) handleLuaDo(msg sarif.Message) {
	machine := strings.TrimLeft(strings.TrimPrefix(msg.Action, "lua/do"), "/")
	if machine == "" {
		machine = "default"
	}
	m, err := s.getOrCreateMachine(machine)
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	var gp interface{}
	msg.DecodePayload(&gp)
	out, err, p := m.Do(msg.Text, gp)
	if err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	reply := sarif.CreateMessage("lua/done", p)
	reply.Text = out
	s.Reply(msg, reply)
}

type MsgMachineStatus struct {
	Machine string `json:"machine,omitempty"`
	Status  string `json:"status,omitempty"`
	Out     string `json:"out,omitempty"`
}

func (p MsgMachineStatus) String() string {
	s := "Machine " + p.Machine + " is " + p.Status + "."
	if p.Out != "" {
		s += "\n\n" + p.Out
	}
	return s
}

func (s *Service) handleLuaStart(msg sarif.Message) {
	name := strings.TrimPrefix(strings.TrimPrefix(msg.Action, "lua/start"), "/")
	if name == "" {
		s.ReplyBadRequest(msg, errors.New("No machine name given!"))
		return
	}

	m, err := s.createMachineFromScript(name)
	if err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	s.Reply(msg, sarif.CreateMessage("lua/started", &MsgMachineStatus{
		name,
		"up",
		m.FlushOut(),
	}))
}

func (s *Service) handleLuaStop(msg sarif.Message) {
	name := strings.TrimPrefix(strings.TrimPrefix(msg.Action, "lua/stop"), "/")
	if name == "" {
		s.ReplyBadRequest(msg, errors.New("No machine name given!"))
		return
	}

	if err := s.destroyMachine(name); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	s.Reply(msg, sarif.CreateMessage("lua/stopped", &MsgMachineStatus{
		name,
		"down",
		"",
	}))
}

type ContentPayload struct {
	Content schema.Content `json:"content"`
}

func (p ContentPayload) Text() string {
	return "This message contains content."
}

func (s *Service) handleLuaLoad(msg sarif.Message) {
	gen := false
	name := strings.TrimPrefix(strings.TrimPrefix(msg.Action, "lua/load"), "/")
	if name == "" {
		name, gen = sarif.GenerateId(), true
	}
	if _, ok := s.Machines[name]; ok {
		s.destroyMachine(name)
	}

	m, err := s.createMachine(name)
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	var ctp ContentPayload
	if err := msg.DecodePayload(&ctp); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	text := msg.Text
	if ctp.Content.Url != "" {
		ct, err := content.Get(ctp.Content)
		if err != nil {
			s.ReplyBadRequest(msg, err)
		}
		text = string(ct.Data)
	}

	var gp interface{}
	msg.DecodePayload(&gp)
	out, err, _ := m.Do(text, gp)
	if err != nil {
		s.ReplyBadRequest(msg, err)
		s.destroyMachine(name)
		return
	}

	if !gen {
		f, err := os.Create(s.cfg.ScriptDir + "/" + name + ".lua")
		if err == nil {
			_, err = f.Write([]byte(text))
			defer f.Close()
		}
		if err != nil {
			s.ReplyInternalError(msg, err)
			s.destroyMachine(name)
			return
		}
	}

	s.Reply(msg, sarif.CreateMessage("lua/loaded", &MsgMachineStatus{
		name,
		"up",
		out,
	}))
}

func (s *Service) handleLuaDump(msg sarif.Message) {
	name := strings.TrimPrefix(strings.TrimPrefix(msg.Action, "lua/dump"), "/")
	if name == "" {
		s.ReplyBadRequest(msg, errors.New("No machine name given!"))
		return
	}

	f, err := os.Open(s.cfg.ScriptDir + "/" + name + ".lua")
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	defer f.Close()
	src, err := ioutil.ReadAll(f)
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	ct := content.PutData([]byte(src))
	ct.PutAction = "lua/load/" + name
	ct.Name = name + ".lua"
	s.Reply(msg, sarif.CreateMessage("lua/dumped", ContentPayload{ct}))
}
