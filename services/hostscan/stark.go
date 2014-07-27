package hostscan

import (
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
)

var Module = core.Module{
	Name:        "hostscan",
	Version:     "1.0",
	NewInstance: NewInstance,
}

func init() {
	core.RegisterModule(Module)
}

type Service struct {
	scan *HostScan
	ctx  *core.Context
}

func NewService(ctx *core.Context) (*Service, error) {
	db := ctx.Database

	SetupSchema(db.Driver(), db.DB)

	s := &Service{
		New(db.DB),
		ctx,
	}
	return s, nil
}

func NewInstance(ctx *core.Context) (core.ModuleInstance, error) {
	s, err := NewService(ctx)
	return s, err
}

func (s *Service) Enable() error {
	time.AfterFunc(5*time.Second, s.scheduledUpdate)
	return s.ctx.Proto.Subscribe("devices/fetch_last_status", s.HandleLastStatus)
}

func (s *Service) Disable() error {
	return nil
}

func (s *Service) scheduledUpdate() {
	hosts, err := s.scan.Update()
	if err != nil {
		s.ctx.Log.Errorln("[hostscan:update] error:", err)
	} else {
		s.ctx.Log.Infoln("[hostscan:update] done:", hosts)
	}
	time.AfterFunc(5*time.Minute, s.scheduledUpdate)
}

func (s *Service) HandleLastStatus(msg proto.Message) {
	if name := msg.PayloadGetString("host"); name != "" {
		host, err := s.scan.LastStatus(name)
		s.ctx.Log.Debugln(host)
		if err != nil {
			s.ctx.Log.Warnln(err)
			return
		}
		s.ctx.Proto.Publish(msg.Reply(proto.Message{
			Action: "devices/last_status",
			Payload: map[string]interface{}{
				"host": host,
				"text": host.String(),
			},
		}))
		return
	}

	hosts, err := s.scan.LastStatusAll()
	s.ctx.Log.Debugln(hosts)
	if err != nil {
		s.ctx.Log.Warnln(err)
		return
	}
	s.ctx.Proto.Publish(msg.Reply(proto.Message{
		Action: "devices/last_status",
		Payload: map[string]interface{}{
			"hosts": hosts,
			"text":  "",
		},
	}))
}
