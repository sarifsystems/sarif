package transport

import (
	neturl "net/url"
	"github.com/xconstruct/stark"
)

type transport struct {
	scheme string
	connect func(url string) (stark.Conn, error)
}

var transports map[string]transport

func Register(scheme string, connect func(url string) (stark.Conn, error)) {
	if transports == nil {
		transports = make(map[string]transport)
	}
	transports[scheme] = transport{scheme, connect}
}

type ErrTransport struct {
	scheme string
}

func (e *ErrTransport) Error() string {
	return "Unknown transport for scheme: " + e.scheme
}

func Connect(url string) (stark.Conn, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}

	transport, ok := transports[u.Scheme]
	if !ok {
		return nil, &ErrTransport{u.Scheme}
	}

	return transport.connect(url)
}
