// Package transport provides a generic interface for various stark connections.
package transport

import (
	neturl "net/url"
	"github.com/xconstruct/stark"
)

// Conn represents a bidirectional connection to the stark network.
type Conn interface {
	Read() (*stark.Message, error)
	Write(*stark.Message) error
	Close() error
}

// A ConnManager accepts multiple connections.
type ConnManager interface {
	Connect(conn Conn)
}

type transport struct {
	scheme string
	connect func(url string) (Conn, error)
}

var transports map[string]transport

// Register registers a new transport protocol, so that it can be used by
// services. A transport is chosen based on the scheme (e.g. "tcp" or "local").
func Register(scheme string, connect func(url string) (Conn, error)) {
	if transports == nil {
		transports = make(map[string]transport)
	}
	transports[scheme] = transport{scheme, connect}
}

// ErrTransport is returned when no transport for the chosen scheme was found.
type ErrTransport struct {
	scheme string
}

func (e *ErrTransport) Error() string {
	return "Unknown transport for scheme: " + e.scheme
}

// Connect creates a new connection based on the URL and choses the right
// transport to use based on the scheme part.
func Connect(url string) (Conn, error) {
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
