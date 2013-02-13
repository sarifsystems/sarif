package net

import (
	"encoding/json"
	"log"
	"net"
	neturl "net/url"
	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/transport"
)

const DEFAULT_PORT = "6171"

func init() {
	transport.Register("tcp", Connect)
	transport.Register("udp", Connect)
	transport.Register("unix", Connect)
}

type NetTransport struct {
	man transport.ConnManager
	proto string
	address string
	ln net.Listener
}

type netConn struct {
	dec *json.Decoder
	enc *json.Encoder
	conn net.Conn
}

func (c *netConn) Read() (*stark.Message, error) {
	msg := stark.NewMessage()
	if err := c.dec.Decode(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *netConn) Write(msg *stark.Message) error {
	return c.enc.Encode(msg)
}

func (c *netConn) Close() error {
	return c.conn.Close()
}

func NewNetTransport(man transport.ConnManager, url string) (*NetTransport, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}
	if _, _, err = net.SplitHostPort(u.Host); err != nil {
		u.Host += ":" + DEFAULT_PORT
	}
	return &NetTransport{man, u.Scheme, u.Host, nil}, nil
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
			nc := &netConn{
				json.NewDecoder(conn),
				json.NewEncoder(conn),
				conn,
			}
			t.man.Connect(nc)
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

func Connect(url string) (transport.Conn, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}
	if _, _, err = net.SplitHostPort(u.Host); err != nil {
		u.Host += ":" + DEFAULT_PORT
	}

	conn, err := net.Dial(u.Scheme, u.Host)
	if err != nil {
		return nil, err
	}

	return &netConn{
		json.NewDecoder(conn),
		json.NewEncoder(conn),
		conn,
	}, nil
}
