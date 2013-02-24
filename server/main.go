package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/natural"
	"github.com/xconstruct/stark/router"
	"github.com/xconstruct/stark/transport/local"
	"github.com/xconstruct/stark/transport/net"
	"github.com/xconstruct/stark/service"
	"github.com/xconstruct/stark/service/xmpp"
	"github.com/xconstruct/stark/service/reminder"
)

func mpdService() {
	s := service.MustConnect("local://", service.Info{
		Name: "mpd",
		Actions: []string{
			"music.play",
			"music.pause",
			"music.stop",
			"music.prev",
			"music.next",
		},
	})
	for {
		msg, err := s.Read()
		if err != nil {
			return
		}
		action := strings.Split(msg.Action, ".")
		if action[0] == "music" {
			exec.Command("mpc", action[1]).Start()
			reply := stark.NewReply(msg)
			reply.Action = "notify.success"
			reply.Message = "done"
			s.Write(reply)
		}
	}
}

func terminalService() {
	s := service.MustConnect("local://", service.Info{
		Name: "intterminal",
	})
	go func() {
		stdin := bufio.NewReader(os.Stdin)
		for {
			cmd, _ := stdin.ReadString('\n')
			cmd = strings.TrimSpace(cmd)

			msg := stark.NewMessage()
			msg.Action = "natural.process"
			msg.Message = cmd
			msg.Destination = "natural"
			s.Write(msg)
		}
	}()

	for {
		msg, err := s.Read()
		if err != nil {
			return
		}
		fmt.Println(msg.Message)
	}
}

func naturalService() {
	s := service.MustConnect("local://", service.Info{
		Name: "natural",
		Actions: []string{"natural.process"},
	})
	for {
		msg, err := s.Read()
		if err != nil {
			return
		}
		if msg.Action != "natural.process" {
			return
		}

		reply, err := natural.Parse(msg.Message)
		if reply == nil {
			reply = stark.NewReply(msg)
			reply.Action = "error"
			reply.Message = "Did not understand: " + err.Error()
			s.Write(reply)
			continue
		}
		reply.Source = s.Name()
		reply.ReplyTo = msg.Source
		s.Write(reply)
	}
}

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

	go terminalService()
	go mpdService()
	go naturalService()

	rm := reminder.NewReminder("local://")
	rm.Start()

	xs, err := xmpp.NewService("local://", getConfigMap("xmpp"))
	xs.Start()

	select{}
}
