// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package location

import (
	"database/sql"
	"fmt"
	"time"
)

type Location struct {
	Id        int64     `json:"-"`
	Timestamp time.Time `json:"timestamp"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Accuracy  float64   `json:"accuracy"`
	Source    string    `json:"source"`
	Address   string    `json:"address,omitempty"`
}

func (l Location) String() string {
	ts := l.Timestamp.Format(time.RFC1123)
	if l.Address != "" {
		return l.Address + " on " + ts
	}
	return fmt.Sprintf("%.4f, %.4f on %s", l.Latitude, l.Longitude, ts)
}

type Database interface {
	Setup() error
	Store(l Location) error
	GetLastLocationInBounds(latMin, latMax, lngMin, lngMax float64) (Location, error)
	GetLastLocationInCircle(l Location) (Location, error)
}

const schema = `
CREATE TABLE IF NOT EXISTS locations (
	id INT(10) NOT NULL AUTO_INCREMENT
	timestamp TIMESTAMP NOT NULL,
	latitude DECIMAL(9,6) NOT NULL,
	longitude DECIMAL(9,6) NOT NULL,
	accuracy FLOAT NOT NULL,
	source VARCHAR(10) NOT NULL,
	PRIMARY KEY (id),
	UNIQUE KEY latitude (latitude,longitude)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`

const schemaSqlite3 = `
CREATE TABLE IF NOT EXISTS locations (
	id INTEGER PRIMARY KEY,
	timestamp TIMESTAMP NOT NULL,
	latitude DECIMAL(9,6) NOT NULL,
	longitude DECIMAL(9,6) NOT NULL,
	accuracy FLOAT NOT NULL,
	source VARCHAR(10) NOT NULL
);
CREATE INDEX IF NOT EXISTS lat_long ON locations (latitude,longitude);
`

type sqlDatabase struct {
	Driver string
	Db     *sql.DB
}

func (d *sqlDatabase) Setup() error {
	var err error
	if d.Driver == "sqlite3" {
		_, err = d.Db.Exec(schemaSqlite3)
	} else {
		_, err = d.Db.Exec(schema)
	}
	return err
}

func (d *sqlDatabase) GetLastLocationInBounds(latMin, latMax, lngMin, lngMax float64) (Location, error) {
	row := d.Db.QueryRow(`
		SELECT id, timestamp, latitude, longitude, accuracy, source FROM locations
		WHERE latitude BETWEEN ? AND ?
		AND longitude BETWEEN ? AND ?
		ORDER BY timestamp DESC
		LIMIT 1
	`, latMin, latMax, lngMin, lngMax)
	var last Location
	err := row.Scan(&last.Id, &last.Timestamp, &last.Latitude, &last.Longitude,
		&last.Accuracy, &last.Source)
	return last, err
}

func (d *sqlDatabase) GetLastLocationInCircle(l Location) (Location, error) {
	row := d.Db.QueryRow(`
		SELECT id, timestamp, latitude, longitude, accuracy, source FROM locations
		WHERE latitude BETWEEN ? - (?/111111) AND ? + (?/111111)
		AND longitude BETWEEN ? - (?/111111*COS(latitude/180*PI())) AND ? + (?/111111*COS(latitude/180*PI()))
		ORDER BY timestamp DESC
		LIMIT 1
	`, l.Latitude, l.Accuracy, l.Latitude, l.Longitude, l.Accuracy, l.Longitude, l.Accuracy)
	var last Location
	err := row.Scan(&last.Id, &last.Timestamp, &last.Latitude, &last.Longitude,
		&last.Accuracy, &last.Source)
	return last, err
}

func (d *sqlDatabase) Store(l Location) error {
	stmt, err := d.Db.Prepare(`INSERT INTO locations (timestamp, latitude, longitude, accuracy, source) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(l.Timestamp, l.Latitude, l.Longitude, l.Accuracy, l.Source)
	return err
}
