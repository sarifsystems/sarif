package router

import (
	"github.com/xconstruct/stark/proto"
)

type Service interface {
	Handle(msg *proto.Message) *proto.Message
}

type ServiceFunc func(msg *proto.Message) *proto.Message

func (f ServiceFunc) Handle (msg *proto.Message) *proto.Message {
	return f(msg)
}

type Router struct {
	Name string
	Services map[string]Service
}

func NewRouter(name string) *Router {
	return &Router{
		name,
		make(map[string]Service),
	}
}

func (r *Router) Handle(msg *proto.Message) *proto.Message {
	path := proto.GetPath(msg)

	next := path.Next()
	if r.Services[next] != nil {
		reply := r.Services[next].Handle(msg)
		if reply != nil {
			reply = r.Handle(reply)
		}
		return reply
	}

	reply := proto.NewReply(msg)
	reply.Source = r.Name
	reply.Message = "Unknown destination: " + msg.Destination
	return reply
}

func (r *Router) AddService(name string, service Service) {
	r.Services[name] = service
}
