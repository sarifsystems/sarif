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

type Geofence struct {
	Id      int64   `json:"-"`
	LatMin  float64 `json:"lat_min"`
	LatMax  float64 `json:"lat_max"`
	LngMin  float64 `json:"lng_min"`
	LngMax  float64 `json:"lng_max"`
	Name    string  `json:"name,omitempty"`
	Address string  `json:"address,omitempty"`
}

func (g *Geofence) GetBounds() []float64 {
	return []float64{g.LatMin, g.LatMax, g.LngMin, g.LngMax}
}

func (g *Geofence) SetBounds(b []float64) {
	g.LatMin, g.LatMax = b[0], b[1]
	g.LngMin, g.LngMax = b[2], b[3]
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

	StoreLocation(l Location) error
	GetLastLocation() (Location, error)
	GetLastLocationInGeofence(g Geofence) (Location, error)
	GetLastLocationInCircle(l Location) (Location, error)

	StoreGeofence(g Geofence) error
	GetGeofencesInLocation(l Location) ([]Geofence, error)
}

var schema = []string{
	`CREATE TABLE IF NOT EXISTS locations (
		id INT(10) NOT NULL AUTO_INCREMENT,
		timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		latitude DECIMAL(9,6) NOT NULL,
		longitude DECIMAL(9,6) NOT NULL,
		accuracy FLOAT NOT NULL,
		source VARCHAR(10) NOT NULL,
		PRIMARY KEY (id),
		UNIQUE KEY latitude (latitude,longitude)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8;`,

	`CREATE TABLE IF NOT EXISTS location_geofences (
		id INT(10) NOT NULL AUTO_INCREMENT,
		lat_min DECIMAL(9,6) NOT NULL,
		lat_max DECIMAL(9,6) NOT NULL,
		lng_min DECIMAL(9,6) NOT NULL,
		lng_max DECIMAL(9,6) NOT NULL,
		name VARCHAR(100) NOT NULL,
		PRIMARY KEY (id),
		UNIQUE KEY bounds (lat_min, lat_max, lng_min, lng_max)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8;`,
}

var schemaSqlite3 = []string{
	`CREATE TABLE IF NOT EXISTS locations (
		id INTEGER PRIMARY KEY,
		timestamp TIMESTAMP NOT NULL,
		latitude DECIMAL(9,6) NOT NULL,
		longitude DECIMAL(9,6) NOT NULL,
		accuracy FLOAT NOT NULL,
		source VARCHAR(10) NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS lat_long ON locations (latitude,longitude)`,

	`CREATE TABLE IF NOT EXISTS location_geofences (
		id INTEGER PRIMARY KEY,
		lat_min DECIMAL(9,6) NOT NULL,
		lat_max DECIMAL(9,6) NOT NULL,
		lng_min DECIMAL(9,6) NOT NULL,
		lng_max DECIMAL(9,6) NOT NULL,
		name VARCHAR(100) NOT NULL
	);`,
	`CREATE INDEX IF NOT EXISTS bounds ON location_geofences (lat_min, lat_max, lng_min, lng_max);`,
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

func (d *sqlDatabase) GetLastLocation() (Location, error) {
	row := d.Db.QueryRow(`
		SELECT id, timestamp, latitude, longitude, accuracy, source FROM locations
		ORDER BY timestamp DESC
		LIMIT 1`)
	var last Location
	err := row.Scan(&last.Id, &last.Timestamp, &last.Latitude, &last.Longitude,
		&last.Accuracy, &last.Source)
	return last, err
}

func (d *sqlDatabase) GetLastLocationInGeofence(g Geofence) (Location, error) {
	row := d.Db.QueryRow(`
		SELECT id, timestamp, latitude, longitude, accuracy, source FROM locations
		WHERE latitude BETWEEN ? AND ?
		AND longitude BETWEEN ? AND ?
		ORDER BY timestamp DESC
		LIMIT 1
	`, g.LatMin, g.LatMax, g.LngMin, g.LngMax)
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

func (d *sqlDatabase) StoreLocation(l Location) error {
	stmt, err := d.Db.Prepare(`INSERT INTO locations (timestamp, latitude, longitude, accuracy, source) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(l.Timestamp, l.Latitude, l.Longitude, l.Accuracy, l.Source)
	return err
}

func (d *sqlDatabase) StoreGeofence(g Geofence) error {
	stmt, err := d.Db.Prepare(`INSERT INTO location_geofences (lat_min, lat_max, lng_min, lng_max, name) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(g.LatMin, g.LatMax, g.LngMin, g.LngMax, g.Name)
	return err
}

func (d *sqlDatabase) GetGeofencesInLocation(l Location) ([]Geofence, error) {
	rows, err := d.Db.Query(`
		SELECT id, lat_min, lat_max, lng_min, lng_max, name FROM location_geofences
		WHERE ? BETWEEN lat_min AND lat_max
		AND ? BETWEEN lng_min AND lng_max
		LIMIT 1
	`, l.Latitude, l.Longitude)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fences := make([]Geofence, 0)
	for rows.Next() {
		var g Geofence
		err := rows.Scan(&g.Id, &g.LatMin, &g.LatMax, &g.LngMin, &g.LngMax, &g.Name)
		if err != nil {
			return nil, err
		}
		fences = append(fences, g)
	}
	return fences, nil
}
