// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service dbus provides access to DBUS notifications and system state.
package dbus

import (
	"github.com/godbus/dbus"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "dbus",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	Log     proto.Logger
	Session *dbus.Conn
	System  *dbus.Conn
	*proto.Client
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Log:    deps.Log,
		Client: deps.Client,
	}
	return s
}

func (s *Service) Enable() (err error) {
	s.System, err = dbus.SystemBus()
	if err != nil {
		return err
	}

	s.Session, err = dbus.SessionBus()
	if err != nil {
		return err
	}

	s.Subscribe("notify", "", s.handleNotify)
	s.Subscribe("poweroff", "", s.handlePowerOff)
	return nil
}

func (s *Service) handleNotify(msg proto.Message) {
	n := NewNotificationObject(s.Session)
	if err := n.Notify(msg.Text, msg.Text); err != nil {
		s.Log.Errorln("[dbus] notify err:", err)
	}
}

func (s *Service) handlePowerOff(msg proto.Message) {
	o := NewLogindObject(s.System)
	if err := o.PowerOff(); err != nil {
		s.Log.Errorln("[dbus] poweroff err:", err)
	}
}
