// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package location

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "location",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	DB     *core.DB
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	DB  Database
	Log proto.Logger
	*proto.Client
}

func NewService(deps *Dependencies) *Service {
	return &Service{
		&sqlDatabase{deps.DB.Driver(), deps.DB.DB},
		deps.Log,
		deps.Client,
	}
}

func (s *Service) Enable() error {
	if err := s.DB.Setup(); err != nil {
		return err
	}

	if err := s.Subscribe("location/update", "", s.handleLocationUpdate); err != nil {
		return err
	}
	if err := s.Subscribe("location/last", "", s.handleLocationLast); err != nil {
		return err
	}
	if err := s.Subscribe("location/fence/create", "", s.handleGeofenceCreate); err != nil {
		return err
	}
	return nil
}

func (s *Service) Disable() error {
	return nil
}

func fenceInSlice(f Geofence, fences []Geofence) bool {
	for _, ff := range fences {
		if ff.Id == f.Id {
			return true
		}
	}
	return false
}

type GeofenceEventPayload struct {
	Location Location `json:"loc"`
	Fence    Geofence `json:"fence"`
	Status   string   `json:"status"`
}

func (m GeofenceEventPayload) String() string {
	return fmt.Sprintf("%s %ss %s.", m.Location.Source, m.Status, m.Fence.Name)
}

func (s *Service) checkGeofences(last, curr Location) {
	lastFences, err := s.DB.GetGeofencesInLocation(last)
	if err != nil {
		s.Log.Errorln("[location] retrieve last fences", err)
	}
	currFences, err := s.DB.GetGeofencesInLocation(curr)
	if err != nil {
		s.Log.Errorln("[location] retrieve curr fences", err)
	}

	for _, g := range lastFences {
		if !fenceInSlice(g, currFences) {
			s.Log.Debugln("[location] geofence leave:", g)
			pl := GeofenceEventPayload{curr, g, "leave"}
			msg := proto.CreateMessage("location/fence/leave/"+g.Name, pl)
			s.Publish(msg)
		}
	}
	for _, g := range currFences {
		if !fenceInSlice(g, lastFences) {
			s.Log.Debugln("[location] geofence enter:", g)
			pl := GeofenceEventPayload{curr, g, "enter"}
			msg := proto.CreateMessage("location/fence/enter/"+g.Name, pl)
			s.Publish(msg)
		}
	}
}

func (s *Service) handleLocationUpdate(msg proto.Message) {
	loc := Location{}
	if err := msg.DecodePayload(&loc); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	if loc.Timestamp.IsZero() {
		loc.Timestamp = time.Now()
	}
	s.Log.Debugln("[location] store update:", loc)

	last, err := s.DB.GetLastLocation()
	if err != nil {
		s.Log.Errorln("[location] retrieve last err", err)
	}

	if err := s.DB.StoreLocation(loc); err != nil {
		s.ReplyInternalError(msg, err)
	}

	if last.Id != 0 {
		s.checkGeofences(last, loc)
	}
}

type locationLastMessage struct {
	Address   string
	Bounds    []float64
	Latitude  float64
	Longitude float64
	Accuracy  float64
}

var MsgNotFound = proto.Message{
	Action: "err/location/notfound",
	Text:   "No matching location found.",
}

var MsgAddressNotFound = proto.Message{
	Action: "err/location/address/notfound",
	Text:   "Requested address could not be found",
}

func (s *Service) queryLocationLast(pl locationLastMessage) proto.Message {
	if pl.Address != "" {
		geo, err := Geocode(pl.Address)
		if err != nil {
			return proto.InternalError(err)
		}
		if len(geo) == 0 {
			return MsgNotFound
		}
		first := geo[0]
		pl.Address = first.Pretty()
		pl.Bounds = first.BoundingBox
	}

	var loc Location
	var err error
	if len(pl.Bounds) == 4 {
		g := Geofence{}
		g.SetBounds(pl.Bounds)
		loc, err = s.DB.GetLastLocationInGeofence(g)
		loc.Address = pl.Address
	} else {
		loc, err = s.DB.GetLastLocation()
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return MsgNotFound
		}
		return proto.InternalError(err)
	}

	reply := proto.Message{Action: "location/found"}
	if err := reply.EncodePayload(loc); err != nil {
		s.Log.Errorln(err)
		return proto.InternalError(err)
	}
	return reply
}

func (s *Service) handleLocationLast(msg proto.Message) {
	var pl locationLastMessage
	if err := msg.DecodePayload(&pl); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	s.Log.Debugln("[location] last loc request:", pl)

	if reply := s.queryLocationLast(pl); reply.Action != "" {
		reply = msg.Reply(reply)
		s.Publish(reply)
	}
}

func (s *Service) handleGeofenceCreate(msg proto.Message) {
	var g Geofence
	if err := msg.DecodePayload(&g); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	if g.Address != "" {
		geo, err := Geocode(g.Address)
		if err != nil {
			s.ReplyBadRequest(msg, err)
			return
		}
		if len(geo) == 0 {
			s.Publish(msg.Reply(MsgAddressNotFound))
			return
		}
		g.SetBounds(geo[0].BoundingBox)
	}
	if g.Name == "" {
		g.Name = proto.GenerateId()
	}

	if err := s.DB.StoreGeofence(g); err != nil {
		s.ReplyInternalError(msg, err)
	}

	reply := proto.Message{Action: "location/fence/created"}
	if err := reply.EncodePayload(g); err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	reply.Text = "Geofence '" + g.Name + "' created."
	s.Publish(reply)
}
