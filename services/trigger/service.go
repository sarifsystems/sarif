// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service trigger provides simple rules to react to messages on the network.
package trigger

import (
	"errors"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/xconstruct/stark/pkg/template"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "trigger",
	Version:     "1.0",
	NewInstance: NewService,
}

type RuleEntity struct {
	Id int64
	Rule
}

func (RuleEntity) TableName() string {
	return "rules"
}

type Rule struct {
	Name   string `json:"name"`
	Action string `json:"action"`
	Reply  string `json:"reply" sql:"type:text"`
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
	createIndizes := !s.DB.HasTable(&RuleEntity{})
	if err := s.DB.AutoMigrate(&RuleEntity{}).Error; err != nil {
		return err
	}
	if createIndizes {
		if err := s.DB.Model(&RuleEntity{}).AddIndex("action", "action").Error; err != nil {
			return err
		}
	}

	if err := s.Subscribe("trigger/new", "", s.handleTriggerNew); err != nil {
		return err
	}

	actions, err := s.getRuleActions()
	if err != nil {
		return err
	}
	for _, action := range actions {
		if err := s.Subscribe(action, "", s.handleTrigger); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) getRuleActions() ([]string, error) {
	var actions []string
	s.DB.Model(&RuleEntity{}).Group("action").Pluck("action", &actions)
	return actions, s.DB.Error
}

func actionParents(action string) []string {
	parts := strings.Split(action, "/")
	pre := ""
	for i, part := range parts {
		full := pre + part
		parts[i] = full
		pre = full + "/"
	}
	return parts
}

func (s *Service) handleTrigger(msg proto.Message) {
	// Find all rules that are triggered by this action.
	var rules []*RuleEntity
	db := s.DB
	for _, a := range actionParents(msg.Action) {
		db = db.Or("action = ?", a)
	}
	if err := db.Find(&rules).Error; err != nil {
		s.Log.Errorln("[trigger] fetch rules:", err)
		return
	}

	for _, r := range rules {
		tpl, err := template.New().Parse(r.Reply)
		if err != nil {
			s.Log.Errorln("[trigger] setup template:", err)
			continue
		}
		reply := proto.Message{}
		if err := tpl.Execute(&reply, template.MessageToData(msg)); err != nil {
			s.Log.Errorln("[trigger] execute template:", err)
			continue
		}
		s.Publish(reply)
	}
}

func (s *Service) handleTriggerNew(msg proto.Message) {
	r := RuleEntity{}
	if err := msg.DecodePayload(&r); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	if r.Action == "" {
		s.ReplyBadRequest(msg, errors.New("expected action that triggers rule"))
		return
	}

	// Test compiling and executing template.
	tpl, err := template.New().Parse(r.Reply)
	if err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	if err := tpl.Execute(&proto.Message{}, nil); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	// Save rule.
	if err := s.DB.Save(&r).Error; err != nil {
		s.ReplyInternalError(msg, err)
		return
	}
	s.Reply(msg, proto.CreateMessage("trigger/created", &r.Rule))

	// Subscribe to action.
	s.Subscribe(r.Action, "", s.handleTrigger)
}
