// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package lastfm

import (
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

type Artist struct {
	Id         int64
	Name       string `sql:"index"`
	Genre      string
	BroadGenre string
}

func (Artist) TableName() string {
	return "lastfm_artists"
}

type sqlDatabase struct {
	DB *gorm.DB
}

type ByDate []Track

func (a ByDate) Len() int           { return len(a) }
func (a ByDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool { return a[i].Time.Before(a[j].Time) }
