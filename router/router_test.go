package router

import (
	"testing"
	"time"
	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/service"
	"github.com/xconstruct/stark/transport/local"
)

func testSendRec(t *testing.T, sender, receiver *service.Service) {
	sent := stark.NewMessage()
	sent.Action = "notify"
	sent.Message = "router test"
	sent.Destination = receiver.Name()

	if err := sender.Write(sent); err != nil {
		t.Fatal(err)
	}

	got, err := receiver.Read()
	if err != nil {
		t.Fatal(err)
	}

	if sent.UUID != got.UUID {
		t.Errorf("wrong message: got %v, expected %v", got, sent)
	}
}

func TestSimpleRoute(t *testing.T) {
	// Setup router
	r := NewRouter("simple")
	local.NewLocalTransport(r, "local://simple")

	// Setup clients
	a := service.MustConnect("local://simple", service.Info{Name: "a"})
	b := service.MustConnect("local://simple", service.Info{Name: "b"})
	c := service.MustConnect("local://simple", service.Info{Name: "c"})

	// Wait for handshakes
	time.Sleep(100 * time.Millisecond)

	// Test sending
	testSendRec(t, a, b)
	testSendRec(t, a, c)
	testSendRec(t, b, a)
	testSendRec(t, b, c)
	testSendRec(t, c, a)
	testSendRec(t, c, b)
}
