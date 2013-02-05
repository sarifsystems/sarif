package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/router"
	"github.com/xconstruct/stark/natural"
)

var r *router.Router

func echoService(msg *proto.Message) *proto.Message {
	reply := proto.NewReply(msg)
	reply.Action = "notify"
	reply.Message = msg.Message
	return reply
}

func mpdService(msg *proto.Message) *proto.Message {
	exec.Command("mpc", msg.Action).Start()
	reply := proto.NewReply(msg)
	reply.Action = "notify"
	reply.Message = "done"
	return reply
}

func terminalService(msg *proto.Message) *proto.Message {
	fmt.Println(msg.Source + ": " + msg.Message)
	return nil
}

func routes() {
	r.AddService("echo", router.ServiceFunc(echoService))
	r.AddService("mpd", router.ServiceFunc(mpdService))
	r.AddService("natural", router.ServiceFunc(natural.Handle))
	r.AddService("terminal", router.ServiceFunc(terminalService))
}

func main() {
	r = router.NewRouter("router")
	routes()

	stdin := bufio.NewReader(os.Stdin)
	for {
		cmd, _ := stdin.ReadString('\n')
		cmd = strings.TrimSpace(cmd)

		msg := proto.NewMessage()
		msg.Action = natural.ACTION_PROCESS
		msg.Message = cmd
		msg.Source = "terminal"
		msg.Destination = "natural"
		reply := r.Handle(msg)
		if reply != nil {
			fmt.Println(reply)
		}
	}
}
