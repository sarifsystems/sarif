package natural

import (
	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/natural"
	"github.com/xconstruct/stark/service"
)

type Natural struct {
	*service.Service
}

func New(url string) *Natural {
	serv := service.MustConnect(url, service.Info{
		Name: "natural",
		Actions: []string{"natural.process"},
	})
	n := &Natural{serv}
	return n
}

func (n *Natural) Handle(msg *stark.Message) (*stark.Message, error) {
	if msg.Action != "natural.process" {
		return nil, nil
	}

	reply, err := natural.Parse(msg.Message)
	if reply == nil {
		reply = stark.NewReply(msg)
		reply.Action = "error"
		reply.Message = "Did not understand: " + err.Error()
		return reply, nil
	}
	reply.Source = n.Name()
	reply.ReplyTo = msg.Source
	return reply, nil
}

func (n *Natural) Start() {
	go n.HandleLoop(n)
}
