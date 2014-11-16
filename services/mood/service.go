// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mood

import (
	"errors"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "mood",
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
	createIndizes := !s.DB.HasTable(&Tag{})
	if s.DB.AutoMigrate(&Record{}, &Tag{}).Error != nil {
		return s.DB.Error
	}
	if createIndizes {
		if s.DB.Model(&Tag{}).AddUniqueIndex("name", "name").Error != nil {
			return s.DB.Error
		}
	}
	if err := s.Subscribe("mood/record", "", s.handleRecord); err != nil {
		return err
	}
	return nil
}

var ErrNoMoodSpecified = errors.New("No mood specified.")

func (s *Service) handleRecord(msg proto.Message) {
	var m Record
	if err := msg.DecodePayload(&m); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	if m.Text == "" {
		m.Text = msg.Text
	}
	if m.Text == "" {
		s.ReplyBadRequest(msg, ErrNoMoodSpecified)
		return
	}
	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now()
	}

	tags := strings.Split(m.Text, ",")
	for _, name := range tags {
		var tag Tag
		tag.Name = strings.TrimSpace(name)
		s.DB.Where(tag).Attrs(Tag{Score: TagScoreUndefined}).FirstOrCreate(&tag)
		if s.DB.Error != nil {
			s.ReplyInternalError(msg, s.DB.Error)
			return
		}
		m.Tags = append(m.Tags, tag)
	}
	m.RecalculateScore()

	s.Log.Infoln("[mood] new mood:", m)
	if s.DB.Save(&m).Error != nil {
		s.ReplyInternalError(msg, s.DB.Error)
		return
	}

	reply := proto.Message{Action: "mood/recorded"}
	if err := reply.EncodePayload(m); err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	reply.Text = "Mood recorded: " + m.Text

	s.Reply(msg, reply)
}
