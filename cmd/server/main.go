package main

import (
	"github.com/xconstruct/stark/client"
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

	client := client.New(client.Config{
		DeviceName:  "server",
		Server:      cfg.Proto.Server,
		Certificate: cfg.Proto.Certificate,
		Key:         cfg.Proto.Key,
		Authority:   cfg.Proto.Authority,
	})
	assert(client.Connect())

	h := hostscan.NewService(db)
	assert(h.Enable(client))

	select {}
}
