package main

import (
	"log"

	"github.com/xconstruct/goconf"

	"github.com/xconstruct/stark/service"

	_ "github.com/xconstruct/stark/transport/local"
	_ "github.com/xconstruct/stark/transport/net"

	"github.com/xconstruct/stark/service/mpd"
	"github.com/xconstruct/stark/service/natural"
	"github.com/xconstruct/stark/service/reminder"
	"github.com/xconstruct/stark/service/terminal"
	"github.com/xconstruct/stark/service/xmpp"
)


type Dialer interface {
	Dial(url string) error
}

type Server interface {
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
	router := service.New(service.Info{Name: "router"})
	router.Handler = service.LogHandler(service.ForwardHandler)

	// Listen on various protocols, as defined in the config
	for url, _ := range cfg.Listeners {
		if err := router.Listen(url); err != nil {
			log.Println(err)
		}
	}

	services := []Dialer{
		terminal.New(),
		mpd.New(),
		natural.New(),
		reminder.New(),
		xmpp.New(cfg.Xmpp),
	}
	for _, s := range services {
		if err := s.Dial("local://"); err != nil {
			log.Println(err)
			continue
		}

		if serv, ok := s.(Server); ok {
			go serv.Serve()
		}
	}

	// Loop
	select {}
}
