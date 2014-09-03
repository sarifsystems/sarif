// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package events

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	"github.com/xconstruct/stark/util"
)

const (
	StatusSingular   = "singular"
	StatusInProgress = "in_progress"
	StatusStarted    = "started"
	StatusEnded      = "ended"
)

type Event struct {
	Id        int64                  `json:"-"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
	Subject   string                 `json:"subject"`
	Verb      string                 `json:"verb"`
	Object    string                 `json:"object"`
	Status    string                 `json:"status"`
	Source    string                 `json:"source"`
	Text      string                 `json:"text"`
	Meta      map[string]interface{} `json:"meta"`
}

func (e Event) String() string {
	if e.Text == "" {
		e.Text = e.Subject + " " + e.Verb + " " + e.Object
	}
	return util.FuzzyTime(e.Timestamp) + " - " + e.Text
}

type Database interface {
	Setup() error
	StoreEvent(e Event) error
	GetLastEvent(filter Event) (Event, error)
}

const schema = `
CREATE TABLE IF NOT EXISTS events (
	id INT(10) NOT NULL AUTO_INCREMENT,
	timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	subject VARCHAR(100) NOT NULL,
	verb VARCHAR(100) NOT NULL,
	object VARCHAR(100) NOT NULL,
	status VARCHAR(30) NOT NULL,
	source VARCHAR(100) NOT NULL,
	text TEXT,
	meta TEXT,
	PRIMARY KEY (id),
	KEY timestamp (timestamp, subject, verb, object, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`

const schemaSqlite3 = `
CREATE TABLE IF NOT EXISTS events (
	id INTEGER PRIMARY KEY,
	timestamp TIMESTAMP NOT NULL,
	subject VARCHAR(100) NOT NULL,
	verb VARCHAR(100) NOT NULL,
	object VARCHAR(100) NOT NULL,
	status VARCHAR(30) NOT NULL,
	source VARCHAR(100) NOT NULL,
	text TEXT,
	meta TEXT
);
CREATE INDEX IF NOT EXISTS timestamp ON events (timestamp, subject, verb, object, status);
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

func (d *sqlDatabase) StoreEvent(e Event) error {
	meta, err := json.Marshal(e.Meta)
	if err != nil {
		return err
	}
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}

	stmt, err := d.Db.Prepare(`
		INSERT INTO events
		(timestamp, subject, verb, object, status, source, text, meta)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(e.Timestamp, e.Subject, e.Verb, e.Object, e.Status,
		e.Source, e.Text, meta)
	return err
}

func (d *sqlDatabase) GetLastEvent(filter Event) (Event, error) {
	cond := make([]string, 0)
	vars := make([]interface{}, 0)

	if filter.Subject != "" {
		cond = append(cond, "subject = ?")
		vars = append(vars, filter.Subject)
	}
	if filter.Verb != "" {
		cond = append(cond, "verb = ?")
		vars = append(vars, filter.Verb)
	}
	if filter.Object != "" {
		cond = append(cond, "object = ?")
		vars = append(vars, filter.Object)
	}
	if filter.Status != "" {
		cond = append(cond, "status = ?")
		vars = append(vars, filter.Status)
	}
	if filter.Source != "" {
		cond = append(cond, "source = ?")
		vars = append(vars, filter.Source)
	}

	conditions := strings.Join(cond, " AND ")
	if conditions == "" {
		conditions = "1"
	}

	row := d.Db.QueryRow(`
		SELECT timestamp, subject, verb, object, status, source, text, meta
		FROM events
		WHERE `+conditions+`
		ORDER BY timestamp DESC
		LIMIT 1`, vars...)

	var e Event
	meta := ""
	err := row.Scan(&e.Timestamp, &e.Subject, &e.Verb, &e.Object, &e.Status,
		&e.Source, &e.Text, &meta)
	if err != nil {
		return e, err
	}

	err = json.Unmarshal([]byte(meta), &e.Meta)
	return e, err
}
