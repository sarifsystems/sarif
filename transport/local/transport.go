package local

import (
	neturl "net/url"
	"errors"
	"github.com/xconstruct/stark/transport"
)

type LocalListener struct {
	id string
	inc chan transport.Conn
}

var listeners map[string]*LocalListener

func init() {
	listeners = make(map[string]*LocalListener)
	transport.Register(transport.Transport{"local", Dial, Listen})
}

func Listen(url string) (transport.Listener, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}
	if u.Host == "" {
		u.Host = "default"
	}

	t := &LocalListener{u.Host, make(chan transport.Conn, 10)}
	listeners[u.Host] = t
	return t, nil
}

func (t *LocalListener) Accept() (transport.Conn, error) {
	return <-t.inc, nil
}

func (t *LocalListener) Close() error {
	return nil
}

func Dial(url string) (transport.Conn, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}
	if u.Host == "" {
		u.Host = "default"
	}

	t := listeners[u.Host]
	if t == nil {
		return nil, errors.New("Unknown local listener: " + u.Host)
	}
	left, right := NewPipe()
	t.inc <- right
	return left, nil
}
