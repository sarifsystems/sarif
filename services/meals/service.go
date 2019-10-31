// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service meals tracks calories and imports them from fddb.
package meals

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sarifsystems/sarif/pkg/util"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
)

var Module = &services.Module{
	Name:        "meals",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	FDDB struct {
		ApiKey   string
		Username string
		Password string
	}
}

type Dependencies struct {
	DB     *gorm.DB
	Config services.Config
	Client sarif.Client
}

type Service struct {
	cfg Config
	DB  *gorm.DB
	sarif.Client
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		DB:     deps.DB,
		Client: deps.Client,
	}
	deps.Config.Get(&s.cfg)
	return s
}

func (s *Service) Enable() error {
	if err := s.DB.AutoMigrate(&Product{}, &Serving{}).Error; err != nil {
		return err
	}

	if s.cfg.FDDB.ApiKey != "" {
		go s.FddbLoop()
	}

	s.Subscribe("meal/product/new", "", s.handleProductNew)
	s.Subscribe("meal/record", "", s.handleServingRecord)
	s.Subscribe("meal/stats", "", s.handleStats)
	s.Subscribe("meal/fetch_fddb", "", s.handleFetchFddb)
	return nil
}

func (s *Service) handleProductNew(msg sarif.Message) {
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

	s.Reply(msg, sarif.CreateMessage("meal/product/created", &p))
}

func (s *Service) handleServingRecord(msg sarif.Message) {
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
	if sv.Product == nil {
		ps, err := s.findProduct(sv.Name)
		if err != nil {
			s.ReplyBadRequest(msg, err)
			return
		}

		if len(ps) > 1 {
			pList := ""
			for _, p := range ps {
				pList += "\n- " + p.Name
			}
			s.ReplyBadRequest(msg, fmt.Errorf("%d products named %s found.%s", len(ps), sv.Name, pList))
			return
		}
		if len(ps) == 1 {
			sv.Product = &ps[0]
		}
	}
	if sv.AmountWeight <= 0 {
		sv.AmountWeight = Weight(sv.Size) * sv.Product.ServingWeight
	}
	if sv.AmountVolume <= 0 {
		sv.AmountVolume = Volume(sv.Size) * sv.Product.ServingVolume
	}
	if sv.AmountWeight <= 0 && sv.AmountVolume <= 0 {
		s.ReplyBadRequest(msg, errors.New("No serving amount specified."))
		return
	}
	if sv.Time.IsZero() {
		sv.Time = time.Now()
	}

	if err := s.DB.Save(&sv).Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	s.Reply(msg, sarif.CreateMessage("meal/serving/recorded", &sv))
}

func (s *Service) findProduct(name string) ([]Product, error) {
	words := strings.Split(name, " ")
	var ps []Product
	if name == "" {
		return ps, errors.New("No name specified.")
	}
	q := s.DB
	for _, w := range words {
		q = q.Where("name LIKE ?", "%"+w+"%")
	}

	if err := q.Find(&ps).Error; err != nil {
		return ps, err
	}

	return ps, nil
}

var sizeNames = map[string]float64{
	"a":   1,
	"an":  1,
	"one": 1,
	"two": 2,
}

func splitSizeName(s string) (float64, string) {
	parts := strings.SplitN(s, " ", 2)
	if len(parts) != 2 {
		return 0, s
	}
	if v, ok := sizeNames[parts[0]]; ok {
		return v, parts[1]
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

func (s *Service) handleStats(msg sarif.Message) {
	f := ServingFilter{
		After:  time.Now().Truncate(24 * time.Hour).Add(5 * time.Hour),
		Before: time.Now().Truncate(24 * time.Hour).Add(29 * time.Hour),
	}
	if err := msg.DecodePayload(&f); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	var servings []*Serving
	if err := s.DB.Preload("Product").Scopes(applyFilter(f)).Order("time ASC").Preload("Product").Find(&servings).Error; err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	var stats ServingStats
	stats.Servings = servings
	for _, sv := range servings {
		stats.Add(sv.Stats())
	}
	s.Reply(msg, sarif.CreateMessage("meal/stats", stats))
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

func (s *Service) handleFetchFddb(msg sarif.Message) {
	day := util.ParseTime(msg.Text, time.Now())
	if err := s.FetchFddb(day); err != nil {
		s.ReplyInternalError(msg, err)
	}
}
