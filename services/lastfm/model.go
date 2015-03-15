// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package lastfm

import (
	"sort"
	"time"

	"github.com/jinzhu/gorm"
)

type Track struct {
	Id     int64
	Artist string
	Album  string
	Title  string
	Time   time.Time `sql:"index"`
}

func (Track) TableName() string {
	return "lastfm_tracks"
}

type Database interface {
	Setup() error

	StoreTracks(ts []Track) error
	GetLastTrack(filter Track) (Track, error)
}

type sqlDatabase struct {
	DB *gorm.DB
}

func (d *sqlDatabase) Setup() error {
	createIndizes := d.DB.HasTable(&Track{})
	if err := d.DB.AutoMigrate(&Track{}).Error; err != nil {
		return err
	}
	if createIndizes {
		return d.DB.Model(&Track{}).AddIndex("album_artist_title", "album", "artist", "title").Error
	}
	return nil
}

type ByDate []Track

func (a ByDate) Len() int           { return len(a) }
func (a ByDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool { return a[i].Time.Before(a[j].Time) }

func (d *sqlDatabase) StoreTracks(ts []Track) error {
	sort.Sort(ByDate(ts))
	tx := d.DB.Begin()
	for _, t := range ts {
		if err := tx.Save(&t).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

func (d *sqlDatabase) GetLastTrack(filter Track) (Track, error) {
	var t Track
	err := d.DB.Where(&filter).Order("time DESC").First(&t).Error
	return t, err
}
