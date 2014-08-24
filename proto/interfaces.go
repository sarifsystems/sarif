package proto

type Handler func(msg Message)
type Publisher func(msg Message) error

type Endpoint interface {
	Publish(msg Message) error
	RegisterHandler(h Handler)
}

type ReverseEndpoint interface {
	Handle(msg Message)
	RegisterPublisher(p Publisher)
}

func Connect(e Endpoint, r ReverseEndpoint) {
	r.RegisterPublisher(e.Publish)
	e.RegisterHandler(r.Handle)
}
