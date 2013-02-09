package local

import (
	neturl "net/url"
	"errors"
	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/router"
	"github.com/xconstruct/stark/transport"
)

type LocalTransport struct {
	rt *router.Router
	id string
}

var transports map[string]*LocalTransport

func init() {
	transports = make(map[string]*LocalTransport)
	transport.Register("local", Connect)
}

func NewLocalTransport(rt *router.Router, url string) (*LocalTransport, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}
	if u.Host == "" {
		u.Host = "default"
	}

	t := &LocalTransport{rt, u.Host}
	transports[u.Host] = t
	return t, nil
}

func (t *LocalTransport) Connect() (stark.Conn, error) {
	left, right := NewPipe()
	t.rt.Connect(left)
	return right, nil
}

func Connect(url string) (stark.Conn, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}
	if u.Host == "" {
		u.Host = "default"
	}

	t := transports[u.Host]
	if t == nil {
		return nil, errors.New("Unknown local transport: " + u.Host)
	}
	return t.Connect()
}
