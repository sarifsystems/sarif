package router

import (
	"testing"
	"github.com/xconstruct/stark/proto"
)

type ReceiverService struct {
	Last *proto.Message
}

func (s *ReceiverService) Handle(msg *proto.Message) {
	s.Last = msg
}

func TestSimpleRoute(t *testing.T) {
	receiver := &ReceiverService{}
	r := NewRouter("router")
	r.AddService("receiver", receiver)

	sent := proto.NewMessage()
	sent.Source = "test"
	sent.Destination = "receiver"
	sent.Message = "Hello, world!"

	r.Handle(sent)

	got := receiver.Last
	if got == nil {
		t.Fatal("receiver: Message not received")
	}
	if got.Message != "Hello, world!" {
		t.Fatalf("receiver: wrong message, got %v", got.Message)
	}
}
