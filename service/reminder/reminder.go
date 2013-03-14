package reminder

import (
	"time"

	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/service"
)

type Reminder struct {
	*service.Service
}

func New() *Reminder {
	s := service.New(service.Info{
		Name: "reminder",
		Actions: []string{"remind.in"},
	})
	r := &Reminder{s}
	s.Handler = r
	return r
}

func (r *Reminder) Handle(msg *stark.Message) *stark.Message {
	if msg.Action != "remind.in" {
		return stark.ReplyUnsupported(msg)
	}

	in, ok := msg.Data["in"].(string)
	if !ok {
		reply := stark.ReplyError(msg, nil)
		reply.Message = "No duration specified"
		return reply
	}
	dur, err := time.ParseDuration(in)
	if err != nil {
		reply := stark.ReplyError(msg, err)
		reply.Message = "Cannot understand duration for reminder"
		return reply
	}

	time.AfterFunc(dur, func() {
		reply := stark.NewReply(msg)
		reply.Action = "notify.remind"
		reply.Message = "Reminder after " + in
		reason, _:= msg.Data["reason"].(string)
		if reason != "" {
			reply.Message += ":" + reason
		}
		r.Write(reply)
	})
	return nil
}
