// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package events

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "events",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	RecordedActions map[string]bool `json:"recorded_actions"`
}

type Dependencies struct {
	DB     *gorm.DB
	Config services.Config
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	cfg services.Config
	DB  *gorm.DB
	Log proto.Logger
	*proto.Client
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		cfg:    deps.Config,
		DB:     deps.DB,
		Log:    deps.Log,
		Client: deps.Client,
	}
	return s
}

func (s *Service) Enable() error {
	createIndizes := !s.DB.HasTable(&Event{})
	if err := s.DB.AutoMigrate(&Event{}).Error; err != nil {
		return err
	}
	if createIndizes {
		if err := s.DB.Model(&Event{}).AddIndex("events_time", "time", "action").Error; err != nil {
			return err
		}
		if err := s.DB.Model(&Event{}).AddIndex("events_action", "action", "time").Error; err != nil {
			return err
		}
	}
	s.Subscribe("event/new", "", s.handleEventNew)
	s.Subscribe("event/last", "", s.handleEventLast)
	s.Subscribe("event/count", "", s.handleEventCount)
	s.Subscribe("event/list", "", s.handleEventList)
	s.Subscribe("event/sum/duration", "", s.handleEventSumDuration)
	s.Subscribe("event/record", "", s.handleEventRecord)
	s.Subscribe("event/aggregate", "", s.handleEventAggregate)

	var cfg Config
	if !s.cfg.Exists() {
		cfg.RecordedActions = map[string]bool{
			"devices/changed":        true,
			"location/cluster/enter": true,
			"location/cluster/leave": true,
			"location/fence/enter":   true,
			"location/fence/leave":   true,
			"mood":                   true,
		}
	}
	s.cfg.Get(&cfg)
	for action, enabled := range cfg.RecordedActions {
		if enabled {
			if err := s.Subscribe(action, "", s.handleEventNew); err != nil {
				return err
			}
		}
	}
	return nil
}

var MessageEventNotFound = proto.Message{
	Action: "event/notfound",
	Text:   "No event found.",
}

func parseDataFromAction(action, prefix string) (s string, v float64, ok bool) {
	v = 1
	action = strings.TrimLeft(strings.TrimPrefix(action, prefix), "/")
	if action == "" {
		return "", v, false
	}

	parts := strings.Split(action, "/")
	if len(parts) == 1 {
		return parts[0], v, true
	}
	var err error
	if v, err = strconv.ParseFloat(parts[len(parts)-1], 64); err == nil {
		parts = parts[0 : len(parts)-1]
	}
	return strings.Join(parts, "/"), v, true
}

type EventFilter struct {
	Event
	After  time.Time `json:"after"`
	Before time.Time `json:"before"`
}

func applyFilter(f EventFilter) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if action := strings.TrimRight(f.Event.Action, "/"); action != "" {
			db = db.Where("(action = ? OR action LIKE ?)", action, action+"/%")
			f.Event.Action = ""
		}
		db = db.Where(&f.Event)
		if !f.After.IsZero() {
			db = db.Where("time > ?", f.After)
		}
		if !f.Before.IsZero() {
			db = db.Where("time < ?", f.Before)
		}
		return db
	}
}

func (s *Service) handleEventLast(msg proto.Message) {
	var filter EventFilter
	if err := msg.DecodePayload(&filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	s.Log.Infoln("[events] get last by filter:", filter)
	var last Event
	if err := s.DB.Scopes(applyFilter(filter)).Order("time desc").First(&last).Error; err != nil {
		if err == gorm.RecordNotFound {
			s.Reply(msg, MessageEventNotFound)
			return
		}
		s.ReplyInternalError(msg, err)
		return
	}
	s.Log.Infoln("[events] found", last)

	reply := proto.Message{Action: "event/found"}
	if err := reply.EncodePayload(last); err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	reply.Text = last.String()

	s.Reply(msg, reply)
}

type aggPayload struct {
	Type   string      `json:"type,omitempty"`
	Filter EventFilter `json:"filter,omitempty"`
	Events []*Event    `json:"events,omitempty"`
	Value  float64     `json:"value"`
}

func (p aggPayload) Text() string {
	switch p.Type {
	case "count":
		return fmt.Sprintf("Found %g events.", p.Value)
	case "list":
		s := fmt.Sprintf("Found %g events.\n", p.Value)
		for _, e := range p.Events {
			s += "- " + e.String() + "\n"
		}
		return strings.TrimRight(s, "\n")
	}
	return fmt.Sprintf("%s(%s) = %g", p.Type, p.Filter.Action, p.Value)
}

func (s *Service) handleEventCount(msg proto.Message) {
	var filter EventFilter
	if err := msg.DecodePayload(&filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	s.Log.Infoln("[events] get count by filter:", filter)
	count := 0
	if err := s.DB.Model(Event{}).Scopes(applyFilter(filter)).Count(&count).Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	s.Log.Infoln("[events] count", count)

	s.Reply(msg, proto.CreateMessage("events/counted", &aggPayload{
		Type:   "count",
		Filter: filter,
		Value:  float64(count),
	}))
}

func (s *Service) handleEventList(msg proto.Message) {
	var filter EventFilter
	if err := msg.DecodePayload(&filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	s.Log.Infoln("[events] list by filter:", filter)
	var events []*Event
	if err := s.DB.Scopes(applyFilter(filter)).Order("time asc").Find(&events).Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	s.Log.Infoln("[events] found", len(events))

	s.Reply(msg, proto.CreateMessage("events/listed", &aggPayload{
		Type:   "list",
		Filter: filter,
		Events: events,
		Value:  float64(len(events)),
	}))
}

type sumDurationPayload struct {
	Durations map[string]float64 `json:"durations,omitempty"`
	Filter    Event              `json:"filter,omitempty"`
}

func (s *Service) handleEventSumDuration(msg proto.Message) {
	var filter EventFilter
	if err := msg.DecodePayload(&filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	s.Log.Infoln("[events] get sum by filter:", filter)
	var events []*Event
	if err := s.DB.Scopes(applyFilter(filter)).Order("time asc").Find(&events).Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	if filter.Before.IsZero() {
		filter.Before = time.Now()
	}

	d := make(map[string]float64)
	fmt.Println(events)
	if len(events) > 0 {
		start := filter.After
		last := "unknown"
		for _, e := range events {
			action := e.Action
			if action == "" {
				action = "unknown"
			}
			if last != action {
				if !start.IsZero() {
					d[last] += e.Time.Sub(start).Seconds()
				}
				last = action
				start = e.Time
			}
		}

		if !start.IsZero() {
			d[last] += filter.Before.Sub(start).Seconds()
		}
	}

	s.Reply(msg, proto.CreateMessage("events/summarized/duration", &sumDurationPayload{d, filter.Event}))
}

func (s *Service) handleEventNew(msg proto.Message) {
	isTargeted := msg.IsAction("event/new")

	var e Event
	e.Text = msg.Text
	e.Time = time.Now()
	e.Value = 1
	if s, v, ok := parseDataFromAction(msg.Action, "event/new"); ok {
		e.Action, e.Value = s, v
	}
	if err := msg.DecodePayload(&e); err != nil && isTargeted {
		s.ReplyBadRequest(msg, err)
		return
	}

	if err := msg.DecodePayload(&e.Meta); err != nil {
		s.Log.Warnln("[events] meta decode error: %v, %v", msg, err)
	}

	s.Log.Infoln("[events] new event:", e)
	if err := s.DB.Save(&e).Error; err != nil {
		s.Log.Errorf("[events] internal error: %v, %v", msg, err)
		return
	}

	if isTargeted {
		reply := proto.Message{Action: "event/created"}
		if err := reply.EncodePayload(e); err != nil {
			s.ReplyInternalError(msg, err)
			return
		}
		reply.Text = "New event: " + e.String()
		s.Reply(msg, reply)
	}
}

type recordPayload struct {
	Action string `json:"action"`
}

func (s *Service) handleEventRecord(msg proto.Message) {
	var p recordPayload
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	if p.Action == "" {
		s.ReplyBadRequest(msg, errors.New("No action specified"))
		return
	}

	var cfg Config
	s.cfg.Get(&cfg)
	if cfg.RecordedActions == nil {
		cfg.RecordedActions = make(map[string]bool)
	}
	if enabled := cfg.RecordedActions[p.Action]; !enabled {
		cfg.RecordedActions[p.Action] = true
		s.cfg.Set(cfg)
		s.Subscribe(p.Action, "", s.handleEventNew)
	}

	s.Log.Infoln("[events] recording action:", p.Action)
	s.Reply(msg, proto.CreateMessage("event/recording", p))
}

func (s *Service) handleEventAggregate(msg proto.Message) {
	var filter EventFilter
	if err := msg.DecodePayload(&filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	var events []*Event
	if err := s.DB.Scopes(applyFilter(filter)).Order("time asc").Find(&events).Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	var v float64
	typ := strings.TrimPrefix(msg.Action, "event/aggregate/")
	switch typ {
	case "sum":
		for _, e := range events {
			v += e.Value
		}
	case "min":
		if len(events) > 0 {
			v = events[0].Value
			for _, e := range events {
				v = math.Min(v, e.Value)
			}
		}
	case "max":
		if len(events) > 0 {
			v = events[0].Value
			for _, e := range events {
				v = math.Max(v, e.Value)
			}
		}
	case "avg":
		for _, e := range events {
			v += e.Value
		}
		v /= float64(len(events))
	}

	s.Reply(msg, proto.CreateMessage("event/aggregated/"+typ, &aggPayload{
		Type:   typ,
		Filter: filter,
		Value:  v,
	}))
}
