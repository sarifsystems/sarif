package main

import (
	"fmt"
	"github.com/xconstruct/stark/conf"
	"github.com/xconstruct/stark/database"
	"github.com/xconstruct/stark/log"
	"github.com/xconstruct/stark/services/hostscan"
)

func assert(err error) {
	if err != nil {
		log.Default.Fatalln(err)
	}
}

func main() {
	cfg, err := conf.ReadDefault()
	assert(err)

	db, err := database.Open(cfg)
	assert(err)
	defer db.Close()

	assert(hostscan.SetupSchema(cfg.Db.Driver, db))
	h := hostscan.New(db)
	hosts, err := h.Update()
	assert(err)
	fmt.Println(hosts)
}
