// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

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
	scan  *HostScan
	ctx   *core.Context
	proto *proto.Client
}

func NewService(ctx *core.Context) (*Service, error) {
	db := ctx.Database

	SetupSchema(db.Driver(), db.DB)

	s := &Service{
		New(db.DB),
		ctx,
		nil,
	}
	return s, nil
}

func NewInstance(ctx *core.Context) (core.ModuleInstance, error) {
	s, err := NewService(ctx)
	return s, err
}

func (s *Service) Enable() error {
	time.AfterFunc(5*time.Second, s.scheduledUpdate)
	s.proto = proto.NewClient("hostscan", s.ctx.Proto)
	return s.proto.Subscribe("devices/fetch_last_status", "", s.HandleLastStatus)
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

type HostRequest struct {
	Host string `json:"host"`
}

func (s *Service) HandleLastStatus(msg proto.Message) {
	if msg.Action != "devices/fetch_last_status" {
		return
	}
	req := HostRequest{}
	msg.DecodePayload(&req)

	if req.Host != "" {
		host, err := s.scan.LastStatus(req.Host)
		s.ctx.Log.Debugln(host)
		if err != nil {
			s.ctx.Log.Warnln(err)
			return
		}
		s.proto.Publish(msg.Reply(proto.CreateMessage("devices/last_status", host)))
		return
	}

	hosts, err := s.scan.LastStatusAll()
	s.ctx.Log.Debugln(hosts)
	if err != nil {
		s.ctx.Log.Warnln(err)
		return
	}
	s.proto.Publish(msg.Reply(proto.CreateMessage("devices/last_status", hosts)))
}
