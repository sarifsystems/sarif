package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/natural"
	"github.com/xconstruct/stark/router"
	"github.com/xconstruct/stark/transports/pipe"
)

func mpdService(p *pipe.Pipe) {
	for {
		msg, err := p.Read()
		if err != nil {
			return
		}
		exec.Command("mpc", msg.Action).Start()

		reply := stark.NewReply(msg)
		reply.Action = "notify"
		reply.Message = "done"
		p.Write(reply)
	}
}

func terminalService(p *pipe.Pipe) {
	go func() {
		stdin := bufio.NewReader(os.Stdin)
		for {
			cmd, _ := stdin.ReadString('\n')
			cmd = strings.TrimSpace(cmd)

			msg := natural.NewMessage("terminal", cmd)
			p.Write(msg)
		}
	}()

	for {
		msg, err := p.Read()
		if err != nil {
			return
		}
		fmt.Println(msg.Message)
	}
}

func naturalService(p *pipe.Pipe) {
	for {
		msg, err := p.Read()
		if err != nil {
			return
		}
		if msg.Action != natural.ACTION_PROCESS {
			return
		}

		reply := natural.Parse(msg.Message)
		if reply == nil {
			reply = stark.NewReply(msg)
			reply.Action = "error"
			reply.Source = "natural"
			reply.Message = "Did not understand: " + msg.Message
			p.Write(reply)
			continue
		}
		reply.Source = "natural"
		reply.ReplyTo = msg.Source
		p.Write(reply)
	}
}

func main() {
	r := router.NewRouter("router")

	left, right := pipe.New()
	go terminalService(left)
	r.Connect("terminal", right)

	left, right = pipe.New()
	go mpdService(left)
	r.Connect("mpd", right)

	left, right = pipe.New()
	go naturalService(left)
	r.Connect("natural", right)

	select{}
}
