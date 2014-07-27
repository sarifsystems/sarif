package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path"
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
