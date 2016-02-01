// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service vdir provides access to VCard contacts and VCal calendars.
package vdir

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
	"github.com/xconstruct/vdir"
)

var Module = &services.Module{
	Name:        "vdir",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	CardDir string
	CalDir  string
}

type Dependencies struct {
	DB     *gorm.DB
	Config services.Config
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	cfg Config
	DB  *gorm.DB
	Log proto.Logger
	*proto.Client

	cards map[string]CardInfo
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		DB:     deps.DB,
		Log:    deps.Log,
		Client: deps.Client,

		cards: make(map[string]CardInfo),
	}
	deps.Config.Get(&s.cfg)
	return s
}

func (s *Service) Enable() error {
	s.Subscribe("vdir/card", "", s.HandleCard)
	if err := s.ReloadFiles(); err != nil {
		s.Log.Errorln("[vdir] reloading files: ", err)
	}
	return nil
}

func (s *Service) ReloadFiles() error {
	if s.cfg.CardDir != "" {
		if err := filepath.Walk(s.cfg.CardDir, s.loadCard); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) loadCard(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	if strings.HasPrefix(info.Name(), ".") {
		return nil
	}
	if !strings.HasSuffix(info.Name(), ".vcf") {
		return nil
	}

	var card Card
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := vdir.NewDecoder(f).Decode(&card); err != nil {
		return err
	}

	card.SchemaContext = "http://www.w3.org/2006/vcard/ns#"
	card.SchemaType = "Individual"
	card.SchemaId = "stark://vdir/card/" + card.Uid
	card.SchemaLabel = card.FormattedName

	ci := CardInfo{
		Id:   card.Uid,
		Path: path,
		Card: &card,
	}
	s.cards[ci.Id] = ci
	return nil
}

func (s *Service) HandleCard(msg proto.Message) {
	uid := strings.TrimPrefix(msg.Action, "vdir/card/")
	c, ok := s.cards[uid]
	if !ok {
		s.ReplyBadRequest(msg, errors.New("No card with with UID "+uid+" found!"))
		return
	}

	s.Reply(msg, proto.CreateMessage("vdir/card", c.Card))
}
