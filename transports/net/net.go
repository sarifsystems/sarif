package net

import (
	"encoding/json"
	"log"
	"net"
	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/router"
)

type NetTransport struct {
	rt *router.Router
	proto string
	address string
	ln net.Listener
}

type NetConn struct {
	dec *json.Decoder
	enc *json.Encoder
	conn net.Conn
}

func (c *NetConn) Read() (*stark.Message, error) {
	msg := stark.NewMessage()
	if err := c.dec.Decode(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *NetConn) Write(msg *stark.Message) error {
	return c.enc.Encode(msg)
}

func NewNetTransport(rt *router.Router, proto, address string) *NetTransport {
	return &NetTransport{rt, proto, address, nil}
}

func (t *NetTransport) Start() error {
	ln, err := net.Listen(t.proto, t.address)
	t.ln = ln
	if err != nil {
		return err
	}

	go func() {
		for {
			conn, err := t.ln.Accept()
			if err != nil {
				log.Printf("net/accept: %v\n", err)
				continue
			}
			nc := &NetConn{
				json.NewDecoder(conn),
				json.NewEncoder(conn),
				conn,
			}
			t.rt.Connect("temp", nc)
		}
	}()

	return nil
}

func (t *NetTransport) Stop() error {
	if t.ln == nil {
		return nil
	}

	err := t.ln.Close()
	t.ln = nil
	return err
}

func Connect(proto, address string) (*NetConn, error) {
	conn, err := net.Dial(proto, address)
	if err != nil {
		return nil, err
	}

	return &NetConn{
		json.NewDecoder(conn),
		json.NewEncoder(conn),
		conn,
	}, nil
}
