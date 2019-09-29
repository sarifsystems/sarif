// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service events stores and manipulates timeseries data.
package events

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
	"github.com/sarifsystems/sarif/services/schema/store"
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
	Config services.Config
	Client sarif.Client
}

type Service struct {
	cfg services.Config
	sarif.Client
	Store *store.Store
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		cfg:    deps.Config,
		Client: deps.Client,
		Store:  store.New(deps.Client),
	}
	return s
}

func (s *Service) Enable() error {
	s.Subscribe("event/new", "", s.handleEventNew)
	s.Subscribe("event/last", "", s.handleEventLast)
	s.Subscribe("event/list", "", s.handleEventList)
	s.Subscribe("event/record", "", s.handleEventRecord)

	var cfg Config
	if !s.cfg.Exists() {
		cfg.RecordedActions = map[string]bool{
			"devices/changed":        true,
			"location/cluster":       true,
			"location/fence":         true,
			"tagged":                 true,
			"browser/session/update": true,
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

var MessageEventNotFound = sarif.Message{
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

func (s *Service) handleEventLast(msg sarif.Message) {
	filter := make(map[string]interface{})
	if err := msg.DecodePayload(&filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	s.Log("debug", "get last by filter:", filter)
	var events []Event
	err := s.Store.Scan("events", store.Scan{
		Reverse: true,
		Only:    "values",
		Filter:  filter,
		Limit:   1,
	}, &events)
	if err != nil {
		s.ReplyInternalError(msg, err)
	}
	if len(events) == 0 {
		s.Reply(msg, MessageEventNotFound)
		return
	}
	s.Log("debug", "last - found", events)

	reply := sarif.Message{Action: "event/found"}
	if err := reply.EncodePayload(events[0]); err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	reply.Text = events[0].String()

	s.Reply(msg, reply)
}

type aggPayload struct {
	Type   string                 `json:"type,omitempty"`
	Filter map[string]interface{} `json:"filter,omitempty"`
	Events []Event                `json:"events,omitempty"`
	Value  float64                `json:"value"`
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
	return fmt.Sprintf("%s = %g", p.Type, p.Value)
}

func hasKey(m map[string]interface{}, key string) bool {
	_, ok := m[key]
	return ok
}

func (s *Service) handleEventList(msg sarif.Message) {
	var filter map[string]interface{}
	if err := msg.DecodePayload(&filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	if filter == nil {
		filter = make(map[string]interface{})
	}
	reverse := false
	if len(filter) == 0 {
		filter["time >="] = time.Now().Add(-24 * time.Hour)
		reverse = true
	}

	s.Log("debug", "list by filter:", filter)
	var events []Event
	err := s.Store.Scan("events", store.Scan{
		Only:    "values",
		Filter:  filter,
		Reverse: reverse,
	}, &events)
	if err != nil {
		s.ReplyInternalError(msg, err)
	}
	s.Log("debug", "list - found", len(events))

	s.Reply(msg, sarif.CreateMessage("events/listed", &aggPayload{
		Type:   "list",
		Filter: filter,
		Events: events,
		Value:  float64(len(events)),
	}))
}

func (s *Service) handleEventNew(msg sarif.Message) {
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
		s.ReplyBadRequest(msg, err)
		return
	}
	if e.Time.IsZero() {
		e.Time = time.Now()
	}

	if _, err := s.Store.Put(e.Key(), &e); err != nil {
		s.Log("err/internal", "could not store finished task: "+err.Error())
		return
	}

	if isTargeted {
		reply := sarif.Message{Action: "event/created"}
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

func (s *Service) handleEventRecord(msg sarif.Message) {
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

	s.Log("debug", "recording action:", p.Action)
	s.Reply(msg, sarif.CreateMessage("event/recording", p))
}
