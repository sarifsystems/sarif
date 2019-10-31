// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service logger logs matching messages to file or stdout.
package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
)

var Module = &services.Module{
	Name:        "logger",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	Actions map[string]string
}

type Dependencies struct {
	Config services.Config
	Client sarif.Client
}

type Service struct {
	sarif.Client
	Cfg    Config
	Config services.Config

	mutex sync.Mutex
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Client: deps.Client,
		Config: deps.Config,
	}
	return s
}

func (s *Service) Enable() error {
	if !s.Config.Exists() {
		dir := s.Config.Dir()
		if dir == "" {
			s.Cfg.Actions = map[string]string{
				"proto/log": "-",
				"log":       "-",
			}
		} else {
			s.Cfg.Actions = map[string]string{
				"proto/log": dir + "/logs/default.log",
				"log":       s.Config.Dir() + "/logs/default.log",
			}
		}
	}
	s.Config.Get(&s.Cfg)
	for action, target := range s.Cfg.Actions {
		s.Subscribe(action, "", s.handleLog)
		if target != "" && target != "-" {
			if err := os.MkdirAll(path.Dir(target), 0700); err != nil {
				return err
			}
		}
	}
	return nil
}

type LogMessage struct {
	Time time.Time `json:"time"`
	sarif.Message
}

func (s *Service) handleLog(msg sarif.Message) {
	targets := make(map[string]struct{})
	for action, target := range s.Cfg.Actions {
		if msg.IsAction(action) {
			targets[target] = struct{}{}
		}
	}

	log.Println(msg.Action, msg.Source, msg.Text)
	lm := LogMessage{time.Now(), msg}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	for target := range targets {
		if target == "" {
			continue
		}
		raw, _ := json.Marshal(lm)
		if err := s.writeTarget(target, string(raw)); err != nil {
			log.Println("[log] write error:", err)
		}
	}
}

func (s *Service) writeTarget(target, out string) error {
	if target == "-" {
		fmt.Println(out)
		return nil
	}

	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(f, out); err != nil {
		return err
	}
	return f.Close()
}
