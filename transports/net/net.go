package net

import (
	"encoding/json"
	"log"
	"github.com/xconstruct/stark"
)

type NetTransport struct {
	rt router.Router
	net string
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

func NewNetTransport(rt router.Router, net, address string) *NetTransport {
	return &NetTransport{rt, service, net, address}
}

func (t *NetTransport) Start() error {
	t.ln, err := net.Listen(t.net, t.address)
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
			rt.Connect("temp", nc)
		}
	}

	return nil
}

func (t *NetTransport) Stop() error {
	if t.ln == nil {
		return
	}

	err := t.ln.Close()
	t.ln = nil
	return err
}

func Connect(net, address string) (*NetConn, error) {
	conn, err := net.Dial(net, address)
	if err != nil {
		return err
	}

	return &NetConn{
		json.NewDecoder(conn),
		json.NewEncoder(conn),
		conn,
	}
}
