// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package location

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "location",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	DB     *gorm.DB
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	DB  *gorm.DB
	Log proto.Logger
	*proto.Client
}

func NewService(deps *Dependencies) *Service {
	return &Service{
		DB:     deps.DB,
		Log:    deps.Log,
		Client: deps.Client,
	}
}

func (s *Service) Enable() error {
	createIndizes := !s.DB.HasTable(&Location{})
	if err := s.DB.AutoMigrate(&Location{}).Error; err != nil {
		return err
	}
	if err := s.DB.AutoMigrate(&Geofence{}).Error; err != nil {
		return err
	}
	if createIndizes {
		if err := s.DB.Model(&Location{}).AddIndex("lat_long", "latitude", "longitude").Error; err != nil {
			return err
		}
		if err := s.DB.Model(&Geofence{}).AddIndex("bounds", "lat_min", "lat_max", "lng_min", "lng_max").Error; err != nil {
			return err
		}
	}
	s.DB.LogMode(true)

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
	var lastFences, currFences []Geofence
	err := s.DB.Where("? BETWEEN lat_min AND lat_max AND ? BETWEEN lng_min AND lng_max", last.Latitude, last.Longitude).Find(&lastFences).Error
	if err != nil {
		s.Log.Errorln("[location] retrieve last fences", err)
	}
	err = s.DB.Where("? BETWEEN lat_min AND lat_max AND ? BETWEEN lng_min AND lng_max", curr.Latitude, curr.Longitude).Find(&currFences).Error
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

	var last Location
	if err := s.DB.Order("timestamp DESC").First(&last).Error; err != nil && err != gorm.RecordNotFound {
		s.Log.Errorln("[location] retrieve last err", err)
	}

	if err := s.DB.Save(&loc).Error; err != nil {
		s.ReplyInternalError(msg, err)
	}

	if last.Id != 0 {
		s.checkGeofences(last, loc)
	}
}

type LocationFilter struct {
	Location
	Bounds []float64
	After  time.Time `json:"after"`
	Before time.Time `json:"before"`
}

func applyFilter(f LocationFilter) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		f.Address = ""
		if !f.After.IsZero() {
			db = db.Where("timestamp > ?", f.After)
		}
		if !f.Before.IsZero() {
			db = db.Where("timestamp < ?", f.Before)
		}
		if len(f.Bounds) == 4 {
			db = db.Where("latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?",
				f.Bounds[0], f.Bounds[1], f.Bounds[2], f.Bounds[3])
		}
		db = db.Where(&f.Location)
		return db
	}
}

var MsgNotFound = proto.Message{
	Action: "err/location/notfound",
	Text:   "No matching location found.",
}

var MsgAddressNotFound = proto.Message{
	Action: "err/location/address/notfound",
	Text:   "Requested address could not be found",
}

func (s *Service) handleLocationLast(msg proto.Message) {
	var pl LocationFilter
	if err := msg.DecodePayload(&pl); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	s.Log.Debugln("[location] last loc request:", pl)

	if pl.Address != "" {
		geo, err := Geocode(pl.Address)
		if err != nil {
			s.ReplyInternalError(msg, err)
			return
		}
		if len(geo) == 0 {
			s.Reply(msg, MsgNotFound)
			return
		}
		first := geo[0]
		pl.Address = first.Pretty()
		pl.Bounds = first.BoundingBox
	}

	var loc Location
	if err := s.DB.Scopes(applyFilter(pl)).Order("timestamp DESC").First(&loc).Error; err != nil {
		if err == gorm.RecordNotFound {
			s.Reply(msg, MsgNotFound)
			return
		}
		s.ReplyInternalError(msg, err)
		return
	}
	if loc.Address == "" {
		loc.Address = pl.Address
	}

	s.Reply(msg, proto.CreateMessage("location/found", loc))
}

type listPayload struct {
	Count     int         `json:"count"`
	Locations []*Location `json:"locations"`
}

func (pl listPayload) Text() string {
	return fmt.Sprintf("Found %d locations.", pl.Count)
}

func (s *Service) handleLocationList(msg proto.Message) {
	var pl LocationFilter
	if err := msg.DecodePayload(&pl); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	s.Log.Debugln("[location] list loc request:", pl)

	if pl.Address != "" {
		geo, err := Geocode(pl.Address)
		if err != nil {
			s.ReplyInternalError(msg, err)
			return
		}
		if len(geo) == 0 {
			s.Reply(msg, MsgNotFound)
			return
		}
		first := geo[0]
		pl.Address = first.Pretty()
		pl.Bounds = first.BoundingBox
	}

	var locs []*Location
	if err := s.DB.Scopes(applyFilter(pl)).Order("timestamp ASC").Limit(300).Find(&locs).Error; err != nil {
		if err == gorm.RecordNotFound {
			s.Reply(msg, MsgNotFound)
			return
		}
		s.ReplyInternalError(msg, err)
		return
	}

	s.Reply(msg, proto.CreateMessage("location/listed", &listPayload{
		len(locs),
		locs,
	}))
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

	if err := s.DB.Save(&g).Error; err != nil {
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
