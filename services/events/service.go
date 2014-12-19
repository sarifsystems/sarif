// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package events

import (
	"fmt"
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
	createIndizes := !s.DB.HasTable(&Event{})
	if err := s.DB.AutoMigrate(&Event{}).Error; err != nil {
		return err
	}
	if createIndizes {
		if err := s.DB.Model(&Event{}).AddIndex("timestamp", "timestamp", "subject", "verb", "verb", "object", "status").Error; err != nil {
			return err
		}
	}
	if err := s.Subscribe("event/new", "", s.handleEventNew); err != nil {
		return err
	}
	if err := s.Subscribe("event/last", "", s.handleEventLast); err != nil {
		return err
	}
	if err := s.Subscribe("event/count", "", s.handleEventCount); err != nil {
		return err
	}
	if err := s.Subscribe("event/sum/duration", "", s.handleEventSumDuration); err != nil {
		return err
	}
	if err := s.Subscribe("location/fence", "", s.handleLocationFence); err != nil {
		return err
	}
	return nil
}

func (s *Service) Disable() error { return nil }

var MessageEventNotFound = proto.Message{
	Action: "event/notfound",
	Text:   "No event found.",
}

func fixEvent(e *Event) {
	if e.Subject == "i" || e.Subject == "I" {
		e.Subject = "user"
	}
}

func (s *Service) handleEventNew(msg proto.Message) {
	var e Event
	if err := msg.DecodePayload(&e); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	fixEvent(&e)
	if e.Text == "" {
		e.Text = msg.Text
	}
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}

	s.Log.Infoln("[events] new event:", e)

	if err := s.DB.Save(&e).Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	reply := proto.Message{Action: "event/created"}
	if err := reply.EncodePayload(e); err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	reply.Text = "New event: " + e.String()

	s.Reply(msg, reply)
}

type EventFilter struct {
	Event
	After  time.Time `json:"after"`
	Before time.Time `json:"before"`
}

func applyFilter(f EventFilter) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db = db.Where(&f.Event)
		if !f.After.IsZero() {
			db = db.Where("timestamp > ?", f.After)
		}
		if !f.Before.IsZero() {
			db = db.Where("timestamp < ?", f.Before)
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
	fixEvent(&filter.Event)

	s.Log.Infoln("[events] get last by filter:", filter)
	var last Event
	s.DB.Scopes(applyFilter(filter)).Order("timestamp desc").First(&last)
	if err := s.DB.Error; err != nil {
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

type countPayload struct {
	Count  int
	Filter Event
}

func (pl countPayload) Text() string {
	return fmt.Sprintf("Found %d events.", pl.Count)
}

func (s *Service) handleEventCount(msg proto.Message) {
	var filter EventFilter
	if err := msg.DecodePayload(&filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	fixEvent(&filter.Event)

	s.Log.Infoln("[events] get count by filter:", filter)
	count := 0
	s.DB.Model(Event{}).Scopes(applyFilter(filter)).Count(&count)
	if err := s.DB.Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	s.Log.Infoln("[events] count", count)

	s.Reply(msg, proto.CreateMessage("action/counted", &countPayload{count, filter.Event}))
}

type sumDurationPayload struct {
	Duration time.Duration
	Filter   Event
}

func (pl sumDurationPayload) Text() string {
	return fmt.Sprintf("The total duration is %s.", pl.Duration)
}

func (s *Service) handleEventSumDuration(msg proto.Message) {
	var filter EventFilter
	if err := msg.DecodePayload(&filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	fixEvent(&filter.Event)

	s.Log.Infoln("[events] get sum by filter:", filter)
	var events []*Event
	s.DB.Scopes(applyFilter(filter)).Order("timestamp asc").Find(&events)
	if err := s.DB.Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	if filter.Before.IsZero() {
		filter.Before = time.Now()
	}

	var d time.Duration
	if len(events) > 0 {
		start := filter.After
		for _, e := range events {
			switch e.Status {
			case StatusStarted:
				start = e.Timestamp
			case StatusEnded:
				if !start.IsZero() {
					d += e.Timestamp.Sub(start)
					start = time.Time{}
				}
			}
		}

		if !start.IsZero() {
			d += filter.Before.Sub(start)
		}
	}

	s.Reply(msg, proto.CreateMessage("action/summarized/duration", &sumDurationPayload{d, filter.Event}))
}
