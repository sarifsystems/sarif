package mpd

import (
	"os/exec"
	"strings"

	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/service"
)

type MPD struct {
	*service.Service
}

func New() *MPD {
	serv := service.New(service.Info{
		Name: "mpd",
		Actions: []string{
			"music.play",
			"music.pause",
			"music.stop",
			"music.prev",
			"music.next",
		},
	})
	m := &MPD{serv}
	serv.Handler = m
	return m
}

func (m *MPD) Handle(msg *stark.Message) *stark.Message {
	action := strings.Split(msg.Action, ".")
	if action[0] != "music" {
		return nil
	}

	if action[1] == "play" {
		artist, _ := msg.Data["artist"].(string)
		if artist != "" {
			exec.Command("mpc", "clear").Run()
			exec.Command("mpc", "findadd", "artist", artist).Run()
		}
	}

	exec.Command("mpc", action[1]).Run()
	reply := stark.NewReply(msg)
	reply.Action = "notify.success"
	reply.Message = "Done."
	return reply
}
