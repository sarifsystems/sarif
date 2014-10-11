// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package router

import (
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "router",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	Config *core.Config
	Log    proto.Logger
}

type Service struct {
	router   *Router
	Log      proto.Logger
	last     Diagnostic
	interval time.Duration
}

func NewService(deps *Dependencies) *Service {
	//db := ctx.Database
	//SetupSchema(db.Driver(), db.DB)

	var cfg Config
	deps.Config.Get("router", &cfg)
	return &Service{
		New(cfg),
		deps.Log,
		Diagnostic{},
		10 * time.Second,
	}
}

func (s *Service) Enable() error {
	err := s.router.Login()
	if err != nil {
		return err
	}
	time.AfterFunc(s.interval*time.Second, s.scheduledUpdate)
	return nil
}

func (s *Service) scheduledUpdate() {
	diag, err := s.router.Diagnostic()
	if err != nil {
		s.Log.Errorln("[router:update] error:", err)
	} else {
		s.Log.Debugf("[router:update] done, interval %v: %v", s.interval, diag)
	}

	speedChanged := diag.DownSpeed != s.last.DownSpeed
	timeReached := time.Since(s.last.Timestamp) > 5*time.Minute
	if speedChanged || timeReached {
		s.last = diag
	}

	if speedChanged {
		s.interval = 10 * time.Second
	} else if s.interval < 5*time.Minute {
		s.interval += s.interval / 2
	}
	time.AfterFunc(s.interval, s.scheduledUpdate)
}
