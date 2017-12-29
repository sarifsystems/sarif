// Copyright (C) 2017 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service scrobbler scrobbles tracks to Last.fm
package scrobbler

import (
	"github.com/sarifsystems/sarif/pkg/schema"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
	"github.com/shkh/lastfm-go/lastfm"
)

var Module = &services.Module{
	Name:        "scrobbler",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	User      string
	Password  string
	ApiKey    string
	ApiSecret string
}

func (c Config) IsValid() bool {
	return c.User != "" && c.Password != "" &&
		c.ApiKey != "" && c.ApiSecret != ""
}

type Dependencies struct {
	Config services.Config
	Client *sarif.Client
}

type Service struct {
	Config Config
	*sarif.Client
	cfg services.Config
	Api *lastfm.Api
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Client: deps.Client,
		cfg:    deps.Config,
	}
	return s
}

func (s *Service) Enable() error {
	s.cfg.Get(&s.Config)
	if !s.Config.IsValid() {
		return nil
	}

	s.Api = lastfm.New(s.Config.ApiKey, s.Config.ApiSecret)
	if err := s.Api.Login(s.Config.User, s.Config.Password); err != nil {
		s.Log("err/internal", err.Error())
		return nil
	}

	s.Subscribe("music/started", "", s.handleNowPlaying)
	s.Subscribe("music/changed", "", s.handleNowPlaying)
	s.Subscribe("music/scrobble", "", s.handleScrobble)

	return nil
}

func (s *Service) handleNowPlaying(msg sarif.Message) {
	var info schema.MusicInfo
	msg.DecodePayload(&info)

	p := infoToParam(info)
	if _, err := s.Api.Track.UpdateNowPlaying(p); err != nil {
		s.Log("err/internal", err.Error())
	}
}

func (s *Service) handleScrobble(msg sarif.Message) {
	var info schema.MusicInfo
	msg.DecodePayload(&info)

	p := infoToParam(info)
	if !info.Time.IsZero() {
		p["timestamp"] = info.Time.Unix()
	}
	if _, err := s.Api.Track.Scrobble(p); err != nil {
		s.Log("err/internal", err.Error())
	}
}

func infoToParam(info schema.MusicInfo) lastfm.P {
	p := lastfm.P{}
	if info.Artist != "" {
		p["artist"] = info.Artist
	}
	if info.Album != "" {
		p["album"] = info.Album
	}
	if info.Track != "" {
		p["track"] = info.Track
	}
	if info.Duration > 0 {
		p["duration"] = info.Duration
	}
	return p
}
