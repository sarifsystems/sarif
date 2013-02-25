package router

import (
	"testing"
	"time"
	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/service"

	"github.com/xconstruct/stark/transport"
	_ "github.com/xconstruct/stark/transport/local"
)

func testSendRec(t *testing.T, sender, receiver *service.Service, path string) {
	sent := stark.NewMessage()
	sent.Action = "notify"
	sent.Message = "router test"
	sent.Destination = path + "" + receiver.Name()

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
	l, _ := transport.Listen("local://simple")
	go r.Listen(l)

	// Setup clients
	a := service.MustConnect("local://simple", service.Info{Name: "a"})
	b := service.MustConnect("local://simple", service.Info{Name: "b"})
	c := service.MustConnect("local://simple", service.Info{Name: "c"})

	// Wait for handshakes
	time.Sleep(100 * time.Millisecond)

	// Test sending
	testSendRec(t, a, b, "")
	testSendRec(t, a, c, "")
	testSendRec(t, b, a, "")
	testSendRec(t, b, c, "")
	testSendRec(t, c, a, "")
	testSendRec(t, c, b, "")
}

func TestMultiRoute(t *testing.T) {
	// Setup routers
	r1 := NewRouter("router1")
	r2 := NewRouter("router2")
	l1, _ := transport.Listen("local://router1")
	l2, _ := transport.Listen("local://router2")
	go r1.Listen(l1)
	go r2.Listen(l2)

	conn, _ := transport.Dial("local://router1")
	r2.Connect(conn)

	// Setup clients
	a := service.MustConnect("local://router1", service.Info{Name: "a"})
	b := service.MustConnect("local://router1", service.Info{Name: "b"})

	c := service.MustConnect("local://router2", service.Info{Name: "c"})
	d := service.MustConnect("local://router2", service.Info{Name: "d"})

	// Wait for handshakes
	time.Sleep(100 * time.Millisecond)

	// Test sending
	testSendRec(t, a, b, "")
	testSendRec(t, a, c, "router2/")
	testSendRec(t, c, a, "router1/")
	testSendRec(t, c, d, "")
}
