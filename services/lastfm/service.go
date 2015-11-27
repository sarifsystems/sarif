// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package lastfm

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "lastfm",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	User   string
	ApiKey string
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

	importing  sync.WaitGroup
	refreshing bool
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Config{},
		deps.DB,
		deps.Log,
		deps.Client,

		sync.WaitGroup{},
		false,
	}
	deps.Config.Get(&s.cfg)
	return s
}

func (s *Service) Enable() error {
	createIndizes := !s.DB.HasTable(&Track{})
	if err := s.DB.AutoMigrate(&Track{}, &Artist{}).Error; err != nil {
		return err
	}
	if createIndizes {
		if err := s.DB.Model(&Track{}).AddIndex("album_artist_title", "album", "artist", "title").Error; err != nil {
			return err
		}
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

	s.Subscribe("lastfm/force_import", "", func(msg proto.Message) {
		if err := s.ImportAll(); err != nil {
			s.ReplyInternalError(msg, err)
		}
	})
	s.Subscribe("lastfm/tag", "", s.handleTag)
	s.Subscribe("cmd/genre", "", s.handleTag)
	return nil
}

func (s *Service) ImportAll() error {
	s.importing.Wait()
	s.importing.Add(1)
	defer s.importing.Done()

	if s.cfg.User == "" {
		return errors.New("No user specified in config!")
	}
	api := NewApi(s.cfg.ApiKey)
	var last Track
	err := s.DB.Order("time DESC").First(&last).Error
	if err != nil && err != gorm.RecordNotFound {
		return err
	}

	initial, err := api.UserGetRecentTracks(s.cfg.User, 1, last.Time.Unix()+60)
	if err != nil {
		return err
	}
	page := initial.RecentTracks.Attr.TotalPages
	for {
		aresult, err := api.UserGetRecentTracks(s.cfg.User, page, last.Time.Unix()+60)
		if err != nil {
			return err
		}
		result := aresult.RecentTracks
		s.Log.Infof("[lastfm] import page %d/%d", result.Attr.Page, result.Attr.TotalPages)
		if len(result.Tracks) == 0 {
			break
		}

		tracks := make([]Track, len(result.Tracks))
		for i, track := range result.Tracks {
			if track.Attr.NowPlaying {
				continue
			}
			if track.Name == "" {
				continue
			}
			tracks[i].Artist = track.Artist.Text
			tracks[i].Album = track.Album.Text
			tracks[i].Title = track.Name
			tracks[i].Time, err = track.ParseDate()
			if err != nil {
				return err
			}
		}

		if err := s.storeTracks(tracks); err != nil {
			return err
		}
		page = result.Attr.Page - 1
		if page == 0 {
			break
		}
	}

	go s.RefreshArtistInfo()

	return nil
}

func (s *Service) storeTracks(ts []Track) error {
	sort.Sort(ByDate(ts))
	tx := s.DB.Begin()
	for _, t := range ts {
		if err := tx.Save(&t).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (s *Service) RefreshArtistInfo() {
	if s.refreshing || s.cfg.ApiKey == "" {
		return
	}
	defer func() {
		s.refreshing = false
	}()
	s.refreshing = true

	api := NewApi(s.cfg.ApiKey)
	for {
		var name string
		err := s.DB.Raw(`
			SELECT t.artist FROM lastfm_tracks t
			LEFT JOIN lastfm_artists a ON a.name = t.artist 
			WHERE a.name IS NULL
			ORDER BY t.time DESC
			LIMIT 1
		`).Row().Scan(&name)
		if err != nil {
			if err != gorm.RecordNotFound {
				s.Log.Errorln("[lastfm] refresh artist err:", err)
			}
			return
		}
		if name == "" {
			return
		}

		tags, err := api.ArtistGetTopTags(name)
		if err != nil {
			s.Log.Errorln("[lastfm] refresh artist api err:", name, err)
			return
		}
		genre, broad := FindGenre(tags.TopTags.Tags)
		artist := &Artist{
			Name:       name,
			Genre:      genre,
			BroadGenre: broad,
		}
		if err = s.DB.Save(artist).Error; err != nil {
			s.Log.Errorln("[lastfm] refresh artist save err:", err)
			return
		}
	}
}

type tagPayload struct {
	Artist     string   `json:"artist"`
	Genre      string   `json:"genre"`
	BroadGenre string   `json:"broad_genre"`
	Tags       []string `json:"tags"`
}

func (p tagPayload) Text() string {
	return fmt.Sprintf("%s is genre %s, %s.", p.Artist, p.Genre, p.BroadGenre)
}

func (s *Service) handleTag(msg proto.Message) {
	var p tagPayload
	p.Artist = msg.Text
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	api := NewApi(s.cfg.ApiKey)
	tags, err := api.ArtistGetTopTags(p.Artist)
	if err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	p.Artist = tags.TopTags.Attr.Artist
	p.Genre, p.BroadGenre = FindGenre(tags.TopTags.Tags)
	for _, t := range tags.TopTags.Tags {
		p.Tags = append(p.Tags, t.Name)
	}
	s.Reply(msg, proto.CreateMessage("lastfm/tagged", p))
}
