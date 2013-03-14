package natural

import (
	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/natural"
	"github.com/xconstruct/stark/service"
)

type Natural struct {
	*service.Service
}

func New() *Natural {
	serv := service.New(service.Info{
		Name: "natural",
		Actions: []string{"natural.process"},
	})
	n := &Natural{serv}
	serv.Handler = n
	return n
}

func (n *Natural) Handle(msg *stark.Message) *stark.Message {
	if msg.Action != "natural.process" {
		return nil
	}

	reply, err := natural.Parse(msg.Message)
	if reply == nil {
		reply = stark.NewReply(msg)
		reply.Action = "error"
		reply.Message = "Did not understand: " + err.Error()
		return reply
	}
	reply.Source = n.Name()
	reply.ReplyTo = msg.Source
	return reply
}
