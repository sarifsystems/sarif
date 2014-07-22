package proto

type MessageHandler func(msg Message)

type Client interface {
	Publish(msg Message) error
	Subscribe(action string, handler MessageHandler) error
}
