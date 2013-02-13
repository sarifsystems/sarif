package reminder

import (
	"time"

	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/service"
)

type Reminder struct {
	serv *service.Service
}

func NewReminder(url string) *Reminder {
	s := service.MustConnect(url, service.Info{
		Name: "reminder",
		Actions: []string{"remind.in"},
	})
	r := &Reminder{s}
	return r
}

func (r *Reminder) Handle(msg *stark.Message) (*stark.Message, error) {
	if msg.Action != "remind.in" {
		return stark.ReplyUnsupported(msg), nil
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
		r.serv.Write(reply)
	})
	return nil, nil
}

func (r *Reminder) Start() {
	go r.serv.HandleLoop(r)
}

