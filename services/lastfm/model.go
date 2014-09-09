// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package lastfm

import (
	"database/sql"
	"strings"
	"time"
)

type Track struct {
	Id     int64
	Artist string
	Album  string
	Title  string
	Time   time.Time
}

var schema = []string{
	`CREATE TABLE IF NOT EXISTS lastfm_tracks (
		id INT(10) NOT NULL AUTO_INCREMENT,
		time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		artist VARCHAR(100) NOT NULL,
		album VARCHAR(100) NOT NULL,
		title VARCHAR(100) NOT NULL,
		PRIMARY KEY (id),
		INDEX time (time),
		INDEX artist_album_title (artist, album, title)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8;`,
}

var schemaSqlite3 = []string{
	`CREATE TABLE IF NOT EXISTS lastfm_tracks (
		id INTEGER PRIMARY KEY,
		time TIMESTAMP NOT NULL,
		artist VARCHAR(100) NOT NULL,
		album VARCHAR(100) NOT NULL,
		title VARCHAR(100) NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS time ON lastfm_tracks (time)`,
	`CREATE INDEX IF NOT EXISTS artist_album_title ON lastfm_tracks (artist, album, title)`,
}

type Database interface {
	Setup() error

	StoreTracks(ts []Track) error
	GetLastTrack(filter Track) (Track, error)
}

type sqlDatabase struct {
	Driver string
	Db     *sql.DB
}

func (d *sqlDatabase) Setup() error {
	var err error
	s := schema
	if d.Driver == "sqlite3" {
		s = schemaSqlite3
	}
	for _, q := range s {
		if _, err = d.Db.Exec(q); err != nil {
			return err
		}
	}
	return err
}

func (d *sqlDatabase) StoreTracks(ts []Track) error {
	stmt, err := d.Db.Prepare(`INSERT INTO lastfm_tracks (time, artist, album, title) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	for _, t := range ts {
		if t.Title == "" {
			continue
		}
		if _, err := stmt.Exec(t.Time, t.Artist, t.Album, t.Title); err != nil {
			return err
		}
	}
	return nil
}

func (d *sqlDatabase) GetLastTrack(filter Track) (Track, error) {
	cond := make([]string, 0)
	vars := make([]interface{}, 0)

	if filter.Artist != "" {
		cond = append(cond, "artist = ?")
		vars = append(vars, filter.Artist)
	}
	if filter.Album != "" {
		cond = append(cond, "album = ?")
		vars = append(vars, filter.Album)
	}
	if filter.Title != "" {
		cond = append(cond, "title = ?")
		vars = append(vars, filter.Title)
	}

	conditions := strings.Join(cond, " AND ")
	if conditions == "" {
		conditions = "1"
	}

	row := d.Db.QueryRow(`
		SELECT id, time, artist, album, title
		FROM lastfm_tracks
		WHERE `+conditions+`
		ORDER BY time DESC
		LIMIT 1`, vars...)

	var t Track
	err := row.Scan(&t.Id, &t.Time, &t.Artist, &t.Album, &t.Title)
	return t, err
}
