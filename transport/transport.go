// Package transport provides a generic interface for various stark connections.
package transport

import (
	neturl "net/url"
	"github.com/xconstruct/stark"
)

type Listener interface {
	Accept() (stark.Conn, error)
	Close() error
}

type DialFunc func(url string) (stark.Conn, error)
type ListenFunc func(url string) (Listener, error)

type Transport struct {
	Scheme string
	Dial DialFunc
	Listen ListenFunc
}

var transports map[string]Transport

// Register registers a new transport protocol, so that it can be used by
// services. A transport is chosen based on the scheme (e.g. "tcp" or "local").
func Register(tp Transport) {
	if transports == nil {
		transports = make(map[string]Transport)
	}
	transports[tp.Scheme] = tp
}

// ErrTransport is returned when no transport for the chosen scheme was found.
type ErrTransport struct {
	scheme string
}

func (e *ErrTransport) Error() string {
	return "Unknown transport for scheme: " + e.scheme
}

// Dial creates a new connection based on the URL and choses the right
// transport to use based on the scheme part.
func Dial(url string) (stark.Conn, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}

	tp, ok := transports[u.Scheme]
	if !ok {
		return nil, &ErrTransport{u.Scheme}
	}

	return tp.Dial(url)
}

func Listen(url string) (Listener, error) {
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}

	tp, ok := transports[u.Scheme]
	if !ok {
		return nil, &ErrTransport{u.Scheme}
	}

	return tp.Listen(url)
}
