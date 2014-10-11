// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"database/sql"
	"os"
	"path"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

type DatabaseConfig struct {
	Driver string
	Source string
}

type Orm struct {
	driver string
	*gorm.DB
}

type DB struct {
	driver string
	*sql.DB
}

func (db *Orm) Driver() string {
	return db.driver
}

func (db *Orm) Database() *DB {
	return &DB{db.driver, db.DB.DB()}
}

func (db *DB) Driver() string {
	return db.driver
}

func OpenDatabase(cfg DatabaseConfig) (*Orm, error) {
	driver := cfg.Driver
	if driver == "sqlite3" {
		err := os.MkdirAll(path.Dir(cfg.Source), 0700)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	sdb, err := gorm.Open(driver, cfg.Source)
	if err != nil {
		return nil, err
	}
	return &Orm{driver, &sdb}, nil
}

func OpenDatabaseInMemory() *Orm {
	sdb, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	return &Orm{"sqlite3", &sdb}
}
