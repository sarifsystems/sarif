// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package database

import (
	"database/sql"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	Driver string
	Source string
}

type DB struct {
	driver string
	*sql.DB
}

func (db *DB) Driver() string {
	return db.driver
}

func Open(cfg Config) (*DB, error) {
	driver := cfg.Driver
	if driver == "sqlite3" {
		err := os.MkdirAll(path.Dir(cfg.Source), 0700)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	sdb, err := sql.Open(driver, cfg.Source)
	if err != nil {
		return nil, err
	}
	return &DB{driver, sdb}, nil
}
