// Copyright (C) 2018 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service js provides JavaScript scripting for the sarif network.
package js

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/sarifsystems/sarif/pkg/content"
	"github.com/sarifsystems/sarif/pkg/schema"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
)

var Module = &services.Module{
	Name:        "js",
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

	Scripts   map[string]string
	Machines  map[string]*Machine
	Listeners map[string][]string
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Broker:    deps.Broker,
		Client:    deps.Client,
		Scripts:   make(map[string]string),
		Machines:  make(map[string]*Machine),
		Listeners: make(map[string][]string),
	}
	s.cfg.ScriptDir = deps.Config.Dir() + "/js"
	deps.Config.Get(&s.cfg)

	s.createMachine("default")
	return s
}

func (s *Service) Enable() error {
	s.Subscribe("js/do", "", s.handleDo)
	s.Subscribe("js/start", "", s.handleStart)
	s.Subscribe("js/stop", "", s.handleStop)
	s.Subscribe("js/status", "", s.handleStatus)
	s.Subscribe("js/put", "", s.handlePut)
	s.Subscribe("js/get", "", s.handleGet)
	s.Subscribe("js/attach", "", s.handleAttach)

	if err := s.readAvailableScripts(); err != nil {
		return err
	}
	for _, f := range s.Scripts {
		s.createMachineFromScript(f)
	}
	return nil
}

func (s *Service) readAvailableScripts() error {
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
		if !strings.HasSuffix(f, ".js") {
			continue
		}
		name := strings.TrimSuffix(f, ".js")
		s.Scripts[name] = f
	}
	return nil
}

func readSource(filename string) ([]byte, error) {
	if filename == "" || filename == "-" {
		return ioutil.ReadAll(os.Stdin)
	}
	return ioutil.ReadFile(filename)
}

func (s *Service) createMachineFromScript(f string) (*Machine, error) {
	s.Log("info", "loading "+f)
	m, err := s.createMachine(strings.TrimSuffix(f, ".js"))
	if err != nil {
		return m, err
	}

	fname := s.cfg.ScriptDir + "/" + f
	_, err = m.Modules.Require(fname, s.cfg.ScriptDir)
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
	m.Modules.AddPath(s.cfg.ScriptDir)
	s.Machines[name] = m
	if listeners, ok := s.Listeners[name]; ok {
		for _, l := range listeners {
			m.Attach(l)
		}
	}
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

func (s *Service) handleDo(msg sarif.Message) {
	machine := strings.TrimLeft(strings.TrimPrefix(msg.Action, "js/do"), "/")
	if machine == "" {
		machine = "default"
	}
	m, err := s.getOrCreateMachine(machine)
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	out, err, p := m.Do(msg.Text)
	if err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	reply := sarif.CreateMessage("js/done", p)
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

func (s *Service) handleStart(msg sarif.Message) {
	name := strings.TrimPrefix(strings.TrimPrefix(msg.Action, "js/start"), "/")
	if name == "" {
		s.ReplyBadRequest(msg, errors.New("No machine name given!"))
		return
	}

	m, err := s.createMachineFromScript(name)
	if err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	s.Reply(msg, sarif.CreateMessage("js/started", &MsgMachineStatus{
		name,
		"up",
		m.FlushOut(),
	}))
}

func (s *Service) handleStop(msg sarif.Message) {
	name := strings.TrimPrefix(strings.TrimPrefix(msg.Action, "js/stop"), "/")
	if name == "" {
		s.ReplyBadRequest(msg, errors.New("No machine name given!"))
		return
	}

	if err := s.destroyMachine(name); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	s.Reply(msg, sarif.CreateMessage("js/stopped", &MsgMachineStatus{
		name,
		"down",
		"",
	}))
}

type MsgMachineAllStatus struct {
	Up     int               `json:"up"`
	Status map[string]string `json:"status"`
}

func (s MsgMachineAllStatus) Text() string {
	return fmt.Sprintf("%d/%d machines running.", s.Up, len(s.Status))
}

func (s *Service) handleStatus(msg sarif.Message) {
	name := msg.ActionSuffix("js/status")

	if name != "" {
		status := "not_found"
		if _, ok := s.Machines[name]; ok {
			status = "up"
		} else if _, ok := s.Scripts[name]; ok {
			status = "down"
		}
		s.Reply(msg, sarif.CreateMessage("js/status", &MsgMachineStatus{
			name,
			status,
			"",
		}))
		return
	}

	status := MsgMachineAllStatus{
		Status: make(map[string]string),
	}
	for name := range s.Scripts {
		if _, ok := s.Machines[name]; ok {
			status.Up++
			status.Status[name] = "up"
		} else {
			status.Status[name] = "down"
		}
	}
	s.Reply(msg, sarif.CreateMessage("js/status", status))
}

type ContentPayload struct {
	Content schema.Content `json:"content"`
}

func (p ContentPayload) Text() string {
	return "This message contains content."
}

func (s *Service) handlePut(msg sarif.Message) {
	gen := false
	name := strings.TrimPrefix(strings.TrimPrefix(msg.Action, "js/put"), "/")
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

	out, err, _ := m.Do(text)
	if err != nil {
		s.ReplyBadRequest(msg, err)
		s.destroyMachine(name)
		return
	}

	if !gen {
		f, err := os.Create(s.cfg.ScriptDir + "/" + name + ".js")
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

	s.Reply(msg, sarif.CreateMessage("js/status", &MsgMachineStatus{
		name,
		"up",
		out,
	}))
}

func (s *Service) handleGet(msg sarif.Message) {
	name := strings.TrimPrefix(strings.TrimPrefix(msg.Action, "js/get"), "/")
	if name == "" {
		s.ReplyBadRequest(msg, errors.New("No machine name given!"))
		return
	}

	f, err := os.Open(s.cfg.ScriptDir + "/" + name + ".js")
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
	ct.PutAction = "js/put/" + name
	ct.Name = name + ".js"
	s.Reply(msg, sarif.CreateMessage("js/script", ContentPayload{ct}))
}

func (s *Service) handleAttach(msg sarif.Message) {
	name := strings.TrimPrefix(strings.TrimPrefix(msg.Action, "js/attach"), "/")
	if name == "" {
		s.ReplyBadRequest(msg, errors.New("No machine name given!"))
		return
	}

	if m, ok := s.Machines[name]; ok {
		m.Attach(msg.Source)
	}

	s.Listeners[name] = append(s.Listeners[name], msg.Source)
	s.Reply(msg, sarif.CreateMessage("js/attached", nil))
}
