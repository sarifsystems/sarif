// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service statenet provides dynamic state change rules.
package statenet

import (
	"fmt"
	"strings"

	"github.com/xconstruct/stark/pkg/petrinet"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "statenet",
	Version:     "1.0",
	NewInstance: NewService,
}

type Transition struct {
	Name string   `json:"name,omitempty"`
	In   []string `json:"in"`
	Out  []string `json:"out"`
}

type Config struct {
	Transitions []Transition `json:"transitions,omitempty"`
}

type Dependencies struct {
	Config services.Config
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	cfg services.Config
	Log proto.Logger
	*proto.Client

	Net     *petrinet.Net
	Trigger *proto.Message
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		cfg:    deps.Config,
		Log:    deps.Log,
		Client: deps.Client,

		Net: petrinet.New(),
	}
	return s
}

func splitClass(s string) (c, r string) {
	i := strings.Index(s, ":")
	if i == -1 {
		return "", s
	}
	return s[0:i], s[i+1:]
}

func isInternal(name string) bool {
	f, _ := splitClass(name)
	return f == "internal" || f == "-"
}

func (s *Service) Enable() error {
	var cfg Config
	s.cfg.Get(&cfg)

	for _, t := range cfg.Transitions {
		s.Net.AddTransition(t.In, t.Out).Name = t.Name
	}

	for _, n := range s.Net.Nodes {
		c, action := splitClass(n.Name)
		if c == "sub" {
			s.Subscribe(action, "", s.handleSpawn)
		}
		if c == "state" || c == "pub" || c == "pubp" {
			n.OnChange = s.publishChange
		}
	}

	return nil
}

func (s *Service) publishChange(name string, prev, curr int) {
	class, action := splitClass(name)

	msg := proto.CreateMessage(action, nil)
	msg.CorrId = s.Trigger.CorrId
	if msg.CorrId == "" {
		msg.CorrId = s.Trigger.Id
	}

	if class == "state" {
		msg.Action = fmt.Sprintf("state/%s/%d", action, curr)
		s.Publish(msg)
		return
	}

	if curr <= prev {
		return
	}
	if class == "pubp" {
		msg.Text = s.Trigger.Text
		msg.Payload = s.Trigger.Payload
	}
	s.Publish(msg)
}

func (s *Service) handleSpawn(msg proto.Message) {
	if msg.Source == s.DeviceId {
		s.Log.Warnln("[petrinet] ignoring own message: " + msg.Action)
		return
	}

	msg.Action = "sub:" + msg.Action
	for _, n := range s.Net.Nodes {
		if msg.IsAction(n.Name) {
			s.Net.Spawn(n.Name, 1)
		}
	}

	s.Trigger = &msg
	if !s.Net.Run(100) {
		s.Log.Errorln("[petrinet] still running")
	}
	s.Trigger = nil
}
