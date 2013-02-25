package main

import (
	"os/exec"
	"strings"

	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/service"
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
