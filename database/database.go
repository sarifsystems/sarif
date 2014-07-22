package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path"

	"github.com/xconstruct/stark/conf"
)

type DB struct {
	driver string
	*sql.DB
}

func (db *DB) Driver() string {
	return db.driver
}

func Open(cfg conf.Config) (*DB, error) {
	driver := cfg.Db.Driver
	if driver == "sqlite3" {
		err := os.MkdirAll(path.Dir(cfg.Db.Source), 0700)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	sdb, err := sql.Open(driver, cfg.Db.Source)
	if err != nil {
		return nil, err
	}
	return &DB{driver, sdb}, nil
}
