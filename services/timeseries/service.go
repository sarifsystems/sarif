// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package timeseries

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

type Point struct {
	Id     int64     `json:"-"`
	Series string    `json:"series,omitempty"`
	Time   time.Time `json:"time,omitempty"`
	Value  float64   `json:"value"`
}

func (p Point) Text() string {
	return fmt.Sprintf("%s on %s is %g", p.Series, p.Time.Format(time.RFC3339), p.Value)
}

var Module = &services.Module{
	Name:        "timeseries",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	RecordedActions map[string]bool `json:"recorded_actions"`
}

type Dependencies struct {
	DB     *gorm.DB
	Config *core.Config
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	cfg *core.Config
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
	createIndizes := !s.DB.HasTable(&Point{})
	if err := s.DB.AutoMigrate(&Point{}).Error; err != nil {
		return err
	}
	if createIndizes {
		if err := s.DB.Model(&Point{}).AddIndex("series", "time").Error; err != nil {
			return err
		}
	}
	s.Subscribe("timeseries/record", "", s.handleRecord)
	s.Subscribe("timeseries/last", "", s.handleLast)
	s.Subscribe("timeseries/count", "", s.handleCount)
	s.Subscribe("timeseries/list", "", s.handleList)

	s.Subscribe("timeseries/sum", "", s.handleAggregate)
	s.Subscribe("timeseries/min", "", s.handleAggregate)
	s.Subscribe("timeseries/max", "", s.handleAggregate)
	s.Subscribe("timeseries/avg", "", s.handleAggregate)

	var cfg Config
	s.cfg.Get("timeseries", &cfg)
	for action, enabled := range cfg.RecordedActions {
		if enabled {
			if err := s.Subscribe(action, "", s.handleRecord); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseDataFromAction(action string) (s string, v float64, ok bool) {
	v = 1
	action = strings.TrimLeft(strings.TrimPrefix(action, "timeseries/record"), "/")
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

func (s *Service) handleRecord(msg proto.Message) {
	var p Point
	p.Time = time.Now()
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	if p.Series == "" {
		if s, v, ok := parseDataFromAction(msg.Action); ok {
			p.Series, p.Value = s, v
		}
	}
	if p.Series == "" {
		s.ReplyBadRequest(msg, errors.New("No series specified"))
		return
	}

	if err := s.DB.Save(&p).Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
}

type PointFilter struct {
	Point
	After  time.Time `json:"after"`
	Before time.Time `json:"before"`
}

func applyFilter(f PointFilter) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		db = db.Where(&f.Point)
		if !f.After.IsZero() {
			db = db.Where("time > ?", f.After)
		}
		if !f.Before.IsZero() {
			db = db.Where("time < ?", f.Before)
		}
		return db
	}
}

func (s *Service) handleLast(msg proto.Message) {
	var filter PointFilter
	if err := msg.DecodePayload(&filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	var last Point
	if err := s.DB.Scopes(applyFilter(filter)).Order("time desc").First(&last).Error; err != nil {
		if err == gorm.RecordNotFound {
			s.Reply(msg, proto.Message{
				Action: "timeseries/notfound",
				Text:   "No timeseries data found.",
			})
			return
		}
		s.ReplyInternalError(msg, err)
		return
	}
	s.Reply(msg, proto.CreateMessage("timeseries/found", last))
}

type aggPayload struct {
	Type   string      `json:"type,omitempty"`
	Filter PointFilter `json:"filter,omitempty"`
	Points []*Point    `json:"points,omitempty"`
	Value  float64     `json:"value"`
}

func (p aggPayload) Text() string {
	return fmt.Sprintf("%s(%s) = %g", p.Type, p.Filter.Series, p.Value)
}

func (s *Service) handleCount(msg proto.Message) {
	var filter PointFilter
	if err := msg.DecodePayload(&filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	count := 0
	if err := s.DB.Model(Point{}).Scopes(applyFilter(filter)).Count(&count).Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	s.Reply(msg, proto.CreateMessage("timeseries/counted", &aggPayload{
		Type:  "count",
		Value: float64(count),
	}))
}

func (s *Service) handleList(msg proto.Message) {
	var filter PointFilter
	if err := msg.DecodePayload(&filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	var points []*Point
	if err := s.DB.Model(Point{}).Scopes(applyFilter(filter)).Find(&points).Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	s.Reply(msg, proto.CreateMessage("timeseries/listed", &aggPayload{
		Type:   "list",
		Points: points,
		Value:  float64(len(points)),
	}))
}

func (s *Service) handleAggregate(msg proto.Message) {
	var filter PointFilter
	if err := msg.DecodePayload(&filter); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	var points []*Point
	if err := s.DB.Model(Point{}).Scopes(applyFilter(filter)).Find(&points).Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	var v float64
	switch strings.TrimPrefix(msg.Action, "timeseries/") {
	case "sum":
		for _, p := range points {
			v += p.Value
		}
	case "min":
		if len(points) > 0 {
			v = points[0].Value
			for _, p := range points {
				v = math.Min(v, p.Value)
			}
		}
	case "max":
		if len(points) > 0 {
			v = points[0].Value
			for _, p := range points {
				v = math.Max(v, p.Value)
			}
		}
	case "avg":
		for _, p := range points {
			v += p.Value
		}
		v /= float64(len(points))
	}

	s.Reply(msg, proto.CreateMessage("timeseries/listed", &aggPayload{
		Type:   strings.TrimPrefix(msg.Action, "timeseries/"),
		Filter: filter,
		Value:  v,
	}))
}
