package main

import (
	"time"

	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/service"
)

func reminderService() {
	s := service.MustConnect("local://", service.Info{
		Name: "reminder",
		Actions: []string{"remind.in"},
	})

	for {
		msg, err := s.Read()
		if err != nil {
			return
		}
		if msg.Action != "remind.in" {
			continue
		}

		in, ok := msg.Data["in"].(string)
		if !ok {
			reply := stark.NewReply(msg)
			reply.Action = "error"
			reply.Message = "No duration specified"
			s.Write(reply)
			continue
		}
		dur, err := time.ParseDuration(in)
		if err != nil {
			reply := stark.NewReply(msg)
			reply.Action = "error"
			reply.Message = "Cannot understand duration for reminder"
			s.Write(reply)
			continue
		}

		time.AfterFunc(dur, func() {
			reply := stark.NewReply(msg)
			reply.Action = "notify.remind"
			reply.Message = "Reminder after " + in
			reason, _:= msg.Data["reason"].(string)
			if reason != "" {
				reply.Message += ":" + reason
			}
			s.Write(reply)
		})
	}
}
