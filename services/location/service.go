// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service location stores the user location and provides automatic checkins and geocoding.
package location

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
	"github.com/sarifsystems/sarif/services/schema/store"
)

var Module = &services.Module{
	Name:        "location",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	Client *sarif.Client
}

type Service struct {
	*sarif.Client
	Store *store.Store

	Clusters *ClusterGenerator
}

func NewService(deps *Dependencies) *Service {
	return &Service{
		Client: deps.Client,
		Store:  store.New(deps.Client),

		Clusters: NewClusterGenerator(),
	}
}

func (s *Service) Enable() error {
	s.Subscribe("location/update", "", s.handleLocationUpdate)
	s.Subscribe("location/last", "", s.handleLocationLast)
	s.Subscribe("location/list", "", s.handleLocationList)
	s.Subscribe("location/fence/create", "", s.handleGeofenceCreate)
	s.Subscribe("location/import", "", s.handleLocationImport)
	return nil
}

func fenceInSlice(f Geofence, fences []Geofence) bool {
	for _, ff := range fences {
		if ff.Name == f.Name {
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
	err := s.Store.Scan("location_geofences", store.Scan{
		Only: "values",
		Filter: map[string]interface{}{
			"lat_min <=": last.Latitude,
			"lat_max >=": last.Latitude,
			"lng_min <=": last.Longitude,
			"lng_max >=": last.Longitude,
		},
	}, &lastFences)
	if err != nil {
		s.Log("err/internal", "retrieve last fences: "+err.Error())
	}
	err = s.Store.Scan("location_geofences", store.Scan{
		Only: "values",
		Filter: map[string]interface{}{
			"lat_min <=": curr.Latitude,
			"lat_max >=": curr.Latitude,
			"lng_min <=": curr.Longitude,
			"lng_max >=": curr.Longitude,
		},
	}, &currFences)
	if err != nil {
		s.Log("err/internal", "retrieve curr fences: "+err.Error())
	}

	for _, g := range lastFences {
		if !fenceInSlice(g, currFences) {
			s.Log("debug", "geofence leave", g)
			pl := GeofenceEventPayload{curr, g, "leave"}
			msg := sarif.CreateMessage("location/fence/leave/"+g.Name, pl)
			s.Publish(msg)
		}
	}
	for _, g := range currFences {
		if !fenceInSlice(g, lastFences) {
			s.Log("debug", "geofence enter", g)
			pl := GeofenceEventPayload{curr, g, "enter"}
			msg := sarif.CreateMessage("location/fence/enter/"+g.Name, pl)
			s.Publish(msg)
		}
	}
}

func (s *Service) handleLocationUpdate(msg sarif.Message) {
	loc := Location{}
	if err := msg.DecodePayload(&loc); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	if loc.Time.IsZero() {
		loc.Time = time.Now()
	}
	loc.Geohash = EncodeGeohash(loc.Latitude, loc.Longitude, 12)
	s.Log("debug", "store update", loc)

	var last []Location
	err := s.Store.Scan("locations", store.Scan{
		Reverse: true,
		Only:    "values",
		Limit:   1,
	}, &last)
	if err != nil {
		s.Log("err/internal", "retrieve last err: "+err.Error())
	}
	if len(last) > 0 {
		loc.Distance = HaversineDistance(last[0], loc)
		loc.Speed = loc.Distance / loc.Time.Sub(last[0].Time).Seconds()
	}

	if _, err := s.Store.Put(loc.Key(), &loc); err != nil {
		s.ReplyInternalError(msg, err)
	}

	if changed := s.Clusters.Advance(loc); changed {
		c := s.Clusters.Current()
		status := "enter"
		if c.Status != ConfirmedCluster {
			status = "leave"
			c = s.Clusters.LastCompleted()
			s.Clusters.ClearCompleted()
		}

		// TODO: make optional
		if place, err := ReverseGeocode(c.Location); err == nil {
			c.Address = place.Pretty()
		}

		s.Publish(sarif.CreateMessage("location/cluster/"+status, c))
	}

	if len(last) > 0 {
		s.checkGeofences(last[0], loc)
	}
}

var MsgNotFound = sarif.Message{
	Action: "err/location/notfound",
	Text:   "No matching location found.",
}

var MsgAddressNotFound = sarif.Message{
	Action: "err/location/address/notfound",
	Text:   "Requested address could not be found",
}

func (s *Service) fixFilters(filter map[string]interface{}) error {
	var bounds BoundingBox
	if v, ok := filter["bounds"]; ok {
		delete(filter, "bounds")
		if bs, ok := v.([]interface{}); ok {
			fs := make([]float64, len(bs))
			for i, f := range bs {
				fs[i] = f.(float64)
			}
			bounds.SetBounds(fs)
		}
	}
	if v, ok := filter["address"]; ok {
		addr := v.(string)
		delete(filter, "address")

		geo, err := Geocode(addr)
		if err != nil {
			return err
		}
		if len(geo) == 0 {
			return errors.New("Requested address could not be found")
		}
		first := geo[0]
		if bounds.LatMin == 0 {
			bounds = BoundingBox(first.BoundingBox)
		}
	}
	// TODO: support for secondary indizes
	if bounds.LatMin != 0 {
		filter["latitude >="] = bounds.LatMin
		filter["latitude <="] = bounds.LatMax
		filter["longitude >="] = bounds.LngMin
		filter["longitude <="] = bounds.LngMax
	}
	return nil
}

func (s *Service) handleLocationLast(msg sarif.Message) {
	filter := make(map[string]interface{})
	if err := msg.DecodePayload(&filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	s.Log("debug", "last loc request", filter)

	if err := s.fixFilters(filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	var last []Location
	err := s.Store.Scan("locations", store.Scan{
		Reverse: true,
		Only:    "values",
		Filter:  filter,
		Limit:   1,
	}, &last)
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	if len(last) == 0 {
		s.Reply(msg, MsgNotFound)
		return
	}
	if last[0].Address == "" {
		// TODO: make optional
		if place, err := ReverseGeocode(last[0]); err == nil {
			last[0].Address = place.Pretty()
		}
	}

	s.Reply(msg, sarif.CreateMessage("location/found", last[0]))
}

type listPayload struct {
	Count     int         `json:"count"`
	Locations []*Location `json:"locations"`
}

func (pl listPayload) Text() string {
	return fmt.Sprintf("Found %d locations.", pl.Count)
}

func (s *Service) handleLocationList(msg sarif.Message) {
	filter := make(map[string]interface{})
	if err := msg.DecodePayload(&filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	if len(filter) == 0 {
		filter["time >="] = time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339Nano)
	}
	s.Log("debug", "list loc request", filter)

	if err := s.fixFilters(filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	var last []*Location
	err := s.Store.Scan("locations", store.Scan{
		Reverse: true,
		Only:    "values",
		Filter:  filter,
	}, &last)
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	if len(last) == 0 {
		s.Reply(msg, MsgNotFound)
		return
	}

	s.Reply(msg, sarif.CreateMessage("location/listed", &listPayload{
		len(last),
		last,
	}))
}

func (s *Service) handleGeofenceCreate(msg sarif.Message) {
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
		g.BoundingBox = BoundingBox(geo[0].BoundingBox)
	}
	if g.Name == "" {
		g.Name = sarif.GenerateId()
	}
	g.GeohashMin = EncodeGeohash(g.BoundingBox.LatMin, g.BoundingBox.LngMin, 12)
	g.GeohashMax = EncodeGeohash(g.BoundingBox.LatMax, g.BoundingBox.LngMax, 12)

	if _, err := s.Store.Put(g.Key(), &g); err != nil {
		s.ReplyInternalError(msg, err)
	}

	reply := sarif.Message{Action: "location/fence/created"}
	if err := reply.EncodePayload(g); err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	reply.Text = "Geofence '" + g.Name + "' created."
	s.Publish(reply)
}

type importPayload struct {
	CSV       string      `json:"csv,omitempty"`
	Locations []*Location `json:"locations,omitempty"`
}

type importedPayload struct {
	StartTime   time.Time `json:"start_time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
	NumImported int       `json:"num_imported"`
	NumIgnored  int       `json:"num_ignored"`
	NumExisting int       `json:"num_existing"`
	NumTotal    int       `json:"num_total"`
}

func (p importedPayload) Text() string {
	return fmt.Sprintf("Imported %d of %d locations.", p.NumImported, p.NumTotal)
}

func (s *Service) handleLocationImport(msg sarif.Message) {
	p := importPayload{}
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	if len(p.CSV) > 0 {
		locs, err := ReadCSV(strings.NewReader(p.CSV))
		if err != nil {
			s.ReplyBadRequest(msg, err)
			return
		}
		p.Locations = append(p.Locations, locs...)
	}
	if len(p.Locations) == 0 {
		s.ReplyBadRequest(msg, errors.New("No location data found!"))
		return
	}

	var minTime, maxTime time.Time
	for i, loc := range p.Locations {
		if loc.Time.IsZero() {
			s.ReplyBadRequest(msg, errors.New(fmt.Sprintf("Location %d has no time.", i)))
			return
		}
		if minTime.IsZero() || loc.Time.Before(minTime) {
			minTime = loc.Time
		}
		if maxTime.IsZero() || loc.Time.After(maxTime) {
			maxTime = loc.Time
		}
		loc.Geohash = EncodeGeohash(loc.Latitude, loc.Longitude, 12)
	}

	var existing []*Location
	err := s.Store.Scan("locations", store.Scan{
		Only: "values",
		Filter: map[string]interface{}{
			"time >=": minTime.Add(-time.Minute),
			"time <=": maxTime.Add(time.Minute),
		},
		Limit: 500,
	}, &existing)
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	missing := make([]*Location, 0, len(p.Locations))
NEXT_LOC:
	for _, loc := range p.Locations {
		for _, ex := range existing {
			d := ex.Time.Sub(loc.Time)
			if d > -time.Minute && d < time.Minute {
				continue NEXT_LOC
			}
		}
		missing = append(missing, loc)
	}

	// TODO: batch insert when store supports it
	for _, loc := range missing {
		if _, err := s.Store.Put(loc.Key(), &loc); err != nil {
			s.ReplyInternalError(msg, err)
			return
		}
	}

	s.Reply(msg, sarif.CreateMessage("location/imported", &importedPayload{
		StartTime:   minTime,
		EndTime:     maxTime,
		NumImported: len(missing),
		NumIgnored:  len(p.Locations) - len(missing),
		NumExisting: len(existing),
		NumTotal:    len(p.Locations),
	}))
}
