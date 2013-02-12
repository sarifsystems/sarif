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

	s.HandleLoop(service.HandleFunc(func(msg *stark.Message) (*stark.Message, error) {
		if msg.Action != "remind.in" {
			return nil, nil
		}

		in, ok := msg.Data["in"].(string)
		if !ok {
			reply := stark.ReplyError(msg, nil)
			reply.Message = "No duration specified"
			return reply, nil
		}
		dur, err := time.ParseDuration(in)
		if err != nil {
			reply := stark.ReplyError(msg, err)
			reply.Message = "Cannot understand duration for reminder"
			return reply, nil
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
		return nil, nil
	}))
}
