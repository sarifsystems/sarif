package local

import (
	neturl "net/url"
	"errors"
	"github.com/xconstruct/stark/transport"
)

type LocalTransport struct {
	man transport.ConnManager
	id string
}

var transports map[string]*LocalTransport

func init() {
	transports = make(map[string]*LocalTransport)
	transport.Register("local", Connect)
}

func NewLocalTransport(man transport.ConnManager, url string) (*LocalTransport, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}
	if u.Host == "" {
		u.Host = "default"
	}

	t := &LocalTransport{man, u.Host}
	transports[u.Host] = t
	return t, nil
}

func (t *LocalTransport) Connect() (transport.Conn, error) {
	left, right := NewPipe()
	t.man.Connect(left)
	return right, nil
}

func Connect(url string) (transport.Conn, error) {
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
