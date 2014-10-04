// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package store

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"strings"
	"time"
)

var (
	ErrNoResult = errors.New("No result found.")
)

type Document struct {
	Key     string    `json:"key"`
	Value   []byte    `json:"value,omitempty"`
	Updated time.Time `json:"created"`
}

func (doc Document) String() string {
	return "Document " + doc.Key + "."
}

type Store interface {
	Setup() error
	Put(doc Document) (Document, error)
	Get(key string) (Document, error)
	Del(key string) error
}

const schema = `
CREATE TABLE IF NOT EXISTS store (
	id INT(10) NOT NULL AUTO_INCREMENT,
	dkey VARCHAR(255) NOT NULL,
	value TEXT,
	updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	KEY dkey (dkey),
	KEY updated (updated)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
`

const schemaSqlite3 = `
CREATE TABLE IF NOT EXISTS store (
	id INTEGER PRIMARY KEY,
	dkey VARCHAR(255) NOT NULL,
	value TEXT,
	updated TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS dkey ON store (dkey);
CREATE INDEX IF NOT EXISTS updated ON store (updated);
`

type sqlStore struct {
	Driver string
	Db     *sql.DB
}

func (d *sqlStore) Setup() error {
	var err error
	if d.Driver == "sqlite3" {
		_, err = d.Db.Exec(schemaSqlite3)
	} else {
		_, err = d.Db.Exec(schema)
	}
	return err
}

func generateId() string {
	const alphanum = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	var bytes = make([]byte, 12)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func (d *sqlStore) Put(doc Document) (Document, error) {
	if doc.Key == "" {
		doc.Key = generateId()
	} else if strings.HasSuffix(doc.Key, "$") {
		doc.Key = strings.TrimSuffix(doc.Key, "$")
		doc.Key += generateId()
	} else {
		if err := d.Del(doc.Key); err != nil {
			return doc, err
		}
	}
	doc.Updated = time.Now()

	stmt, err := d.Db.Prepare(`
		INSERT INTO store
		(dkey, value, updated)
		VALUES (?, ?, ?)`)
	if err != nil {
		return doc, err
	}
	_, err = stmt.Exec(doc.Key, doc.Value, doc.Updated)
	return doc, err
}

func (d *sqlStore) Del(key string) error {
	if key == "" {
		return nil
	}
	_, err := d.Db.Exec("DELETE FROM store WHERE dkey = ?", key)
	return err
}

func (d *sqlStore) Get(key string) (Document, error) {
	row := d.Db.QueryRow(`
		SELECT dkey, value, updated
		FROM store
		WHERE dkey = ?
		LIMIT 1`, key)
	var doc Document
	err := row.Scan(&doc.Key, &doc.Value, &doc.Updated)
	if err == sql.ErrNoRows {
		return doc, ErrNoResult
	}
	return doc, err
}
