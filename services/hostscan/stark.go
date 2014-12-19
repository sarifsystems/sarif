// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package hostscan

import (
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "hostscan",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	DB     *core.DB
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	scan *HostScan
	Log  proto.Logger
	*proto.Client
}

func NewService(deps *Dependencies) *Service {
	SetupSchema(deps.DB.Driver(), deps.DB.DB)

	return &Service{
		New(deps.DB.DB),
		deps.Log,
		deps.Client,
	}
}

func (s *Service) Enable() error {
	time.AfterFunc(5*time.Minute, s.scheduledUpdate)
	if err := s.Subscribe("devices/force_update", "", s.HandleForceUpdate); err != nil {
		return err
	}
	return s.Subscribe("devices/fetch_last_status", "", s.HandleLastStatus)
}

func (s *Service) scheduledUpdate() {
	s.Update()
	time.AfterFunc(5*time.Minute, s.scheduledUpdate)
}

func (s *Service) Update() ([]Host, error) {
	hosts, err := s.scan.Update()
	if err != nil {
		s.Log.Errorln("[hostscan:update] error:", err)
		return hosts, err
	}

	s.Log.Infoln("[hostscan:update] done:", hosts)
	for _, host := range hosts {
		name := host.Name
		if name == "" {
			name = host.Ip
		}
		s.Publish(proto.CreateMessage("devices/changed/"+name+"/"+host.Status, &host))
	}
	return hosts, nil
}

type HostRequest struct {
	Host string `json:"host"`
}

func (s *Service) HandleForceUpdate(msg proto.Message) {
	changed, err := s.Update()
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	s.Reply(msg, proto.CreateMessage("devices/changed", changed))
}

func (s *Service) HandleLastStatus(msg proto.Message) {
	req := HostRequest{}
	msg.DecodePayload(&req)

	if req.Host != "" {
		host, err := s.scan.LastStatus(req.Host)
		s.Log.Debugln(host)
		if err != nil {
			s.Log.Warnln(err)
			return
		}
		s.Reply(msg, proto.CreateMessage("devices/last_status", host))
		return
	}

	hosts, err := s.scan.LastStatusAll()
	s.Log.Debugln(hosts)
	if err != nil {
		s.Log.Warnln(err)
		return
	}
	s.Reply(msg, proto.CreateMessage("devices/last_status", hosts))
}
