package proto

type Handler func(msg Message)
type Publisher func(msg Message) error

type Endpoint interface {
	Publish(msg Message) error
	RegisterHandler(h Handler)
}
