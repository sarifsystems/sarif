// Package net provides a transport over tcp, udp or unix sockets.
package net

import (
	"encoding/json"
	"net"
	neturl "net/url"
	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/transport"
)

// The default port for the stark network.
const DEFAULT_PORT = "6171"

func init() {
	transport.Register(transport.Transport{"tcp", Dial, Listen})
	transport.Register(transport.Transport{"udp", Dial, Listen})
	transport.Register(transport.Transport{"unix", Dial, Listen})
}

type NetListener struct {
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

func Listen(url string) (transport.Listener, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}
	if _, _, err = net.SplitHostPort(u.Host); err != nil {
		u.Host += ":" + DEFAULT_PORT
	}
	ln, err := net.Listen(u.Scheme, u.Host)
	if err != nil {
		return nil, err
	}
	return &NetListener{u.Scheme, u.Host, ln}, nil
}

func (t *NetListener) Accept() (stark.Conn, error) {
	conn, err := t.ln.Accept()
	if err != nil {
		return nil, err
	}
	return &netConn{
		json.NewDecoder(conn),
		json.NewEncoder(conn),
		conn,
	}, nil
}

func (t *NetListener) Close() error {
	return t.ln.Close()
}

func Dial(url string) (stark.Conn, error) {
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
