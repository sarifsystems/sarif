package router

import (
	"time"

	"github.com/xconstruct/stark/core"
)

var Module = core.Module{
	Name:        "router",
	Version:     "1.0",
	NewInstance: NewInstance,
}

func init() {
	core.RegisterModule(Module)
}

type Service struct {
	router   *Router
	ctx      *core.Context
	last     Diagnostic
	interval time.Duration
}

func NewService(ctx *core.Context) (*Service, error) {
	//db := ctx.Database
	//SetupSchema(db.Driver(), db.DB)

	var cfg Config
	if err := ctx.Config.Get("router", &cfg); err != nil {
		return nil, err
	}
	s := &Service{
		New(cfg),
		ctx,
		Diagnostic{},
		10 * time.Second,
	}
	return s, nil
}

func NewInstance(ctx *core.Context) (core.ModuleInstance, error) {
	s, err := NewService(ctx)
	return s, err
}

func (s *Service) Enable() error {
	err := s.router.Login()
	if err != nil {
		return err
	}
	time.AfterFunc(s.interval*time.Second, s.scheduledUpdate)
	return nil
}

func (s *Service) Disable() error {
	return nil
}

func (s *Service) scheduledUpdate() {
	diag, err := s.router.Diagnostic()
	if err != nil {
		s.ctx.Log.Errorln("[router:update] error:", err)
	} else {
		s.ctx.Log.Debugf("[router:update] done, interval %v: %v", s.interval, diag)
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