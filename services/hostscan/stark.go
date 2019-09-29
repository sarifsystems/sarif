// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package hostscan

import (
	"time"

	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
	"github.com/sarifsystems/sarif/sfproto"
)

var Module = &services.Module{
	Name:        "hostscan",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	Log    sfproto.Logger
	Client sarif.Client
}

type Service struct {
	scan *HostScan
	Log  sfproto.Logger
	sarif.Client
}

func NewService(deps *Dependencies) *Service {
	return &Service{
		New(),
		deps.Log,
		deps.Client,
	}
}

func (s *Service) Enable() error {
	time.AfterFunc(5*time.Minute, s.scheduledUpdate)
	s.scan.MinDownInterval = 7 * time.Minute
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
		s.Publish(sarif.CreateMessage("devices/changed/"+name+"/"+host.Status, &host))
	}
	return hosts, nil
}

type HostRequest struct {
	Host string `json:"host"`
}

func (s *Service) HandleForceUpdate(msg sarif.Message) {
	changed, err := s.Update()
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	s.Reply(msg, sarif.CreateMessage("devices/changed", changed))
}

func (s *Service) HandleLastStatus(msg sarif.Message) {
	req := HostRequest{}
	msg.DecodePayload(&req)

	if req.Host != "" {
		host, err := s.scan.LastStatus(req.Host)
		s.Log.Debugln(host)
		if err != nil {
			s.Log.Warnln(err)
			return
		}
		s.Reply(msg, sarif.CreateMessage("devices/last_status", host))
		return
	}

	hosts, err := s.scan.LastStatusAll()
	s.Log.Debugln(hosts)
	if err != nil {
		s.Log.Warnln(err)
		return
	}
	s.Reply(msg, sarif.CreateMessage("devices/last_status", hosts))
}
