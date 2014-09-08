// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package scheduler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/util"
)

type Task struct {
	Id         int64         `json:"-"`
	Duration   string        `json:"duration,omitempty"`
	Time       time.Time     `json:"time,omitempty"`
	Location   string        `json:"location,omitempty"`
	Reply      proto.Message `json:"reply,omitempty"`
	CreatedOn  time.Time     `json:"created,omitempty"`
	FinishedOn time.Time     `json:"finished,omitempty"`
}

func (t Task) String() string {
	text := t.Reply.PayloadGetString("text")
	if text == "" {
		text = t.Reply.Action
	}
	return fmt.Sprintf("Schedule task '%s' on %s.",
		text,
		util.FuzzyTime(t.Time),
	)
}

type Database interface {
	Setup() error
	StoreTask(t Task) error
	GetNextTask() (Task, error)
}

const schema = `
CREATE TABLE IF NOT EXISTS scheduler_tasks (
	id INT(10) NOT NULL AUTO_INCREMENT,
	time DATETIME NOT NULL,
	location VARCHAR(100) NOT NULL,
	reply TEXT,
	created_on TIMESTAMP NULL,
	finished_on TIMESTAMP NULL,
	PRIMARY KEY (id),
	KEY time (time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`

const schemaSqlite3 = `
CREATE TABLE IF NOT EXISTS scheduler_tasks (
	id INTEGER PRIMARY KEY,
	time DATETIME NOT NULL,
	location VARCHAR(100) NOT NULL,
	reply TEXT,
	created_on TIMESTAMP NULL,
	finished_on TIMESTAMP NULL
);
CREATE INDEX IF NOT EXISTS time ON scheduler_tasks (time);
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

func (d *sqlDatabase) StoreTask(t Task) error {
	reply, err := json.Marshal(t.Reply)
	if err != nil {
		return err
	}
	if t.CreatedOn.IsZero() {
		t.CreatedOn = time.Now()
	}

	if t.Id != 0 {
		stmt, err := d.Db.Prepare(`
			UPDATE scheduler_tasks
			SET time = ?, location = ?, reply = ?,
				created_on = ?, finished_on = ?
			WHERE id = ?`)
		if err != nil {
			return err
		}
		_, err = stmt.Exec(t.Time, t.Location, string(reply), t.CreatedOn, t.FinishedOn, t.Id)
	} else {
		stmt, err := d.Db.Prepare(`
			INSERT INTO scheduler_tasks
			(time, location, reply, created_on, finished_on)
			VALUES (?, ?, ?, ?, ?)`)
		if err != nil {
			return err
		}
		_, err = stmt.Exec(t.Time, t.Location, reply, t.CreatedOn, t.FinishedOn)
	}
	return err
}

func (d *sqlDatabase) GetNextTask() (Task, error) {
	row := d.Db.QueryRow(`
		SELECT id, time, location, reply, created_on, finished_on
		FROM scheduler_tasks
		WHERE finished_on < DATE('0001-01-02')
		ORDER BY time ASC
		LIMIT 1`)

	var t Task
	reply := ""
	if err := row.Scan(&t.Id, &t.Time, &t.Location, &reply, &t.CreatedOn, &t.FinishedOn); err != nil {
		return t, err
	}

	err := json.Unmarshal([]byte(reply), &t.Reply)
	return t, err
}
