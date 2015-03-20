// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package meals

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "meals",
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
	s := &Service{
		DB:     deps.DB,
		Log:    deps.Log,
		Client: deps.Client,
	}
	return s
}

func (s *Service) Enable() error {
	if err := s.DB.AutoMigrate(&Product{}, &Serving{}).Error; err != nil {
		return err
	}

	s.Subscribe("meal/product/new", "", s.handleProductNew)
	s.Subscribe("meal/record", "", s.handleServingRecord)
	s.Subscribe("meal/stats", "", s.handleStats)
	return nil
}

func (s *Service) handleProductNew(msg proto.Message) {
	var p Product
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	if p.Name == "" {
		s.ReplyBadRequest(msg, errors.New("Please specify a product name."))
		return
	}
	if p.ServingWeight == 0 && p.ServingVolume == 0 {
		s.ReplyBadRequest(msg, errors.New("Please specify a serving size."))
		return
	}

	if err := s.DB.Save(&p).Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	s.Reply(msg, proto.CreateMessage("meal/product/created", &p))
}

func (s *Service) handleServingRecord(msg proto.Message) {
	var sv Serving
	if err := msg.DecodePayload(&sv); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	size, name := splitSizeName(msg.Text)
	if sv.Size == 0 {
		sv.Size = size
	}
	if sv.Name == "" {
		sv.Name = name
	}
	if sv.Size == 0 {
		s.ReplyBadRequest(msg, errors.New("No serving size specified."))
		return
	}
	if sv.Product == nil {
		ps, err := s.findProduct(sv.Name)
		if err != nil {
			s.ReplyBadRequest(msg, err)
			return
		}

		if len(ps) == 0 {
			s.ReplyBadRequest(msg, errors.New("No product named "+name+" found."))
			return
		}
		if len(ps) > 1 {
			s.ReplyBadRequest(msg, fmt.Errorf("%d products named %s found.", len(ps), sv.Name))
			return
		}
		sv.Product = &ps[0]
	}
	if sv.Time.IsZero() {
		sv.Time = time.Now()
	}

	if err := s.DB.Save(&sv).Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	s.Reply(msg, proto.CreateMessage("meal/serving/recorded", &sv))
}

func (s *Service) findProduct(name string) ([]Product, error) {
	var ps []Product
	if name == "" {
		return ps, errors.New("No name specified.")
	}

	if err := s.DB.Where("name LIKE ?", "%"+name+"%").Find(&ps).Error; err != nil {
		return ps, err
	}

	return ps, nil
}

func splitSizeName(s string) (float64, string) {
	parts := strings.SplitN(s, " ", 2)
	if len(parts) != 2 {
		return 0, s
	}
	if parts[0] == "a" || parts[0] == "an" {
		return 1, parts[1]
	}
	if v, err := strconv.ParseFloat(parts[0], 64); err == nil {
		return v, parts[1]
	}
	return 0, s
}

type ServingFilter struct {
	After  time.Time `json:"after"`
	Before time.Time `json:"before"`
}

type ServingStats struct {
	Servings []*Serving `json:"servings,omitempty"`
	Stats
}

func (s ServingStats) String() string {
	return fmt.Sprintf("%d servings totalling %v.", len(s.Servings), s.Stats.Energy.StringKcal())
}

func (s *Service) handleStats(msg proto.Message) {
	f := ServingFilter{
		After:  time.Now().Truncate(24 * time.Hour).Add(5 * time.Hour),
		Before: time.Now().Add(1 * time.Minute),
	}
	if err := msg.DecodePayload(&f); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	var servings []*Serving
	if err := s.DB.Scopes(applyFilter(f)).Order("time ASC").Preload("Product").Find(&servings).Error; err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	var stats ServingStats
	stats.Servings = servings
	for _, sv := range servings {
		stats.Add(sv.Stats())
	}
	s.Reply(msg, proto.CreateMessage("meal/stats", stats))
}

func applyFilter(f ServingFilter) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if !f.After.IsZero() {
			db = db.Where("time > ?", f.After)
		}
		if !f.Before.IsZero() {
			db = db.Where("time < ?", f.Before)
		}
		return db
	}
}
