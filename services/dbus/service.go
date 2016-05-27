// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service dbus provides access to DBUS notifications and system state.
package dbus

import (
	"github.com/godbus/dbus"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
)

var Module = &services.Module{
	Name:        "dbus",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	Log    sarif.Logger
	Client *sarif.Client
}

type Service struct {
	Log     sarif.Logger
	Session *dbus.Conn
	System  *dbus.Conn
	*sarif.Client

	Players map[string]*MprisPlayer
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Log:    deps.Log,
		Client: deps.Client,

		Players: make(map[string]*MprisPlayer),
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

	return s.setupSignals()
}

func (s *Service) handleNotify(msg sarif.Message) {
	n := NewNotificationObject(s.Session)
	if err := n.Notify(msg.Text, msg.Text); err != nil {
		s.Log.Errorln("[dbus] notify err:", err)
	}
}

func (s *Service) handlePowerOff(msg sarif.Message) {
	o := NewLogindObject(s.System)
	if err := o.PowerOff(); err != nil {
		s.Log.Errorln("[dbus] poweroff err:", err)
	}
}

func (s *Service) setupSignals() error {
	c := s.Session.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, "type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged',path='/org/mpris/MediaPlayer2'")
	if c.Err != nil {
		return c.Err
	}

	ch := make(chan *dbus.Signal, 10)
	s.Session.Signal(ch)
	go s.handleSignals(ch)
	return nil
}

func (s *Service) handleSignals(ch chan *dbus.Signal) {
	for v := range ch {
		switch v.Name {
		case "org.freedesktop.DBus.Properties.PropertiesChanged":
			props := v.Body[1].(map[string]dbus.Variant)
			p := s.getPlayer(v.Sender)
			p.UpdateProperties(props)
		}
	}
}

func (s *Service) getPlayer(id string) *MprisPlayer {
	p, ok := s.Players[id]
	if !ok {
		p = &MprisPlayer{Id: id}
		s.Players[id] = p

		obj := s.Session.Object(id, "/org/mpris/MediaPlayer2")
		v, err := obj.GetProperty("org.mpris.MediaPlayer2.Identity")
		if err != nil {
			s.Log.Errorln(err)
		} else {
			p.Identity = v.Value().(string)
		}
		v, err = obj.GetProperty("org.mpris.MediaPlayer2.DesktopEntry")
		if err != nil {
			s.Log.Errorln(err)
		} else {
			p.Name = v.Value().(string)
		}
	}

	return p
}
