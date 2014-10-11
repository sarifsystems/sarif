// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package lastfm

import (
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "lastfm",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	User string
}

type Dependencies struct {
	DB     *core.DB
	Config *core.Config
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	cfg Config
	DB  Database
	Log proto.Logger
	*proto.Client
	importing sync.WaitGroup
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Config{},
		&sqlDatabase{deps.DB.Driver(), deps.DB.DB},
		deps.Log,
		deps.Client,
		sync.WaitGroup{},
	}
	deps.Config.Get("lastfm", &s.cfg)
	return s
}

func (s *Service) Enable() error {
	if err := s.DB.Setup(); err != nil {
		return err
	}

	if s.cfg.User != "" {
		go func() {
			for _ = range time.Tick(30 * time.Minute) {
				if err := s.ImportAll(); err != nil {
					s.Log.Errorln("[lastfm] import err:", err)
				}
			}
		}()
	}
	return nil
}

func (s *Service) ImportAll() error {
	s.importing.Wait()
	s.importing.Add(1)
	defer s.importing.Done()

	if s.cfg.User == "" {
		return errors.New("No user specified in config!")
	}
	api := NewApi()
	last, err := s.DB.GetLastTrack(Track{})
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	page := 100000
	for {
		result, err := api.UserGetRecentTracks(s.cfg.User, page, last.Time.Unix()+60)
		if err != nil {
			return err
		}
		if len(result.Tracks) == 0 {
			return nil
		}
		s.Log.Infof("[lastfm] import page %d/%d", result.Page, result.TotalPages)

		tracks := make([]Track, len(result.Tracks))
		for i, track := range result.Tracks {
			if track.NowPlaying {
				continue
			}
			tracks[i].Artist = track.Artist
			tracks[i].Album = track.Album
			tracks[i].Title = track.Name
			tracks[i].Time, err = track.ParseDate()
			if err != nil {
				return err
			}
		}

		if err := s.DB.StoreTracks(tracks); err != nil {
			return err
		}
		page = result.Page - 1
		if page == 0 {
			return nil
		}
	}

	return nil
}
