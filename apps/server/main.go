package main

import (
	"log"

	"github.com/xconstruct/goconf"

	"github.com/xconstruct/stark/router"
	"github.com/xconstruct/stark/transport"

	_ "github.com/xconstruct/stark/transport/local"
	_ "github.com/xconstruct/stark/transport/net"

	"github.com/xconstruct/stark/service/mpd"
	"github.com/xconstruct/stark/service/natural"
	"github.com/xconstruct/stark/service/reminder"
	"github.com/xconstruct/stark/service/terminal"
	"github.com/xconstruct/stark/service/xmpp"
)


type Service interface {
	Dial(url string) error
	Serve() error
}

type Config struct {
	Listeners map[string]bool
	Xmpp xmpp.Config
}

func main() {
	// Read config
	var cfg Config
	ctx := conf.Build().JSON().Create()
	if err := ctx.Read(&cfg); err != nil {
		log.Fatalln(err)
	}

	// Setup router
	router := router.NewRouter("router")

	// Listen on various protocols, as defined in the config
	for url, _ := range cfg.Listeners {
		listener, err := transport.Listen(url)
		if err != nil {
			log.Println(err)
		}
		go router.Listen(listener)
	}

	services := []Service{
		terminal.New(),
		mpd.New(),
		natural.New(),
		reminder.New(),
		xmpp.New(cfg.Xmpp),
	}
	for _, s := range services {
		err := s.Dial("local://")
		if err != nil {
			log.Println(err)
			continue
		}

		go s.Serve()
	}

	// Loop
	select {}
}
