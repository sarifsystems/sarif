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
)

func mpdService(p *local.Pipe) {
	s, err := service.New("mpd", p)
	if err != nil {
		log.Fatalf("mpc: %v\n", err)
	}
	for {
		msg, err := s.Read()
		if err != nil {
			return
		}
		exec.Command("mpc", msg.Action).Start()

		reply := stark.NewReply(msg)
		reply.Action = "notify"
		reply.Message = "done"
		s.Write(reply)
	}
}

func terminalService(p *local.Pipe) {
	s, err := service.New("terminal", p)
	if err != nil {
		log.Fatalf("intterminal: %v\n", err)
	}
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

func naturalService(p *local.Pipe) {
	s, err := service.New("natural", p)
	if err != nil {
		log.Fatalf("natural: %v\n", err)
	}
	for {
		msg, err := s.Read()
		if err != nil {
			return
		}
		if msg.Action != "natural.process" {
			return
		}

		reply := natural.Parse(msg.Message)
		if reply == nil {
			reply = stark.NewReply(msg)
			reply.Action = "error"
			reply.Message = "Did not understand: " + msg.Message
			p.Write(reply)
			continue
		}
		reply.Source = "natural"
		reply.ReplyTo = msg.Source
		s.Write(reply)
	}
}

func main() {
	r := router.NewRouter("router")

	left, right := local.NewPipe()
	go terminalService(left)
	r.Connect(right)

	left, right = local.NewPipe()
	go mpdService(left)
	r.Connect(right)

	left, right = local.NewPipe()
	go naturalService(left)
	r.Connect(right)

	nt := net.NewNetTransport(r, "tcp", ":9000")
	if err := nt.Start(); err != nil {
		log.Fatalf("server: %v\n", err)
	}

	select{}
}
