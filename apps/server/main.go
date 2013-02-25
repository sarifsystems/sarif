package main

import (
	"log"

	"github.com/xconstruct/stark/router"

	"github.com/xconstruct/stark/transport/local"
	"github.com/xconstruct/stark/transport/net"

	"github.com/xconstruct/stark/service/mpd"
	"github.com/xconstruct/stark/service/natural"
	"github.com/xconstruct/stark/service/reminder"
	"github.com/xconstruct/stark/service/terminal"
	"github.com/xconstruct/stark/service/xmpp"
)

func main() {
	r := router.NewRouter("router")
	local.NewLocalTransport(r, "local://")

	nt, err := net.NewNetTransport(r, "tcp://")
	if err != nil {
		log.Fatalf("server: %v\n", err)
	}
	if err := nt.Start(); err != nil {
		log.Fatalf("server: %v\n", err)
	}

	t := terminal.New("local://")
	t.Start()

	m := mpd.New("local://")
	m.Start()

	n := natural.New("local://")
	n.Start()

	rm := reminder.New("local://")
	rm.Start()

	xs, err := xmpp.NewService("local://", getConfigMap("xmpp"))
	xs.Start()

	select{}
}
