package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path"

	"github.com/xconstruct/stark/conf"
)

func Open(cfg conf.Config) (*sql.DB, error) {
	if cfg.Db.Driver == "sqlite3" {
		err := os.MkdirAll(path.Dir(cfg.Db.Source), 0700)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}
	return sql.Open(cfg.Db.Driver, cfg.Db.Source)
}
