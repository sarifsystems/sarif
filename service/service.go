// Package service provides a wrapper class to create services without the need
// to care about the underlying connection.
//
// The Service class handles such things as connecting and identifying to the router and
// broadcasting the capabilities for you.
package service

import (
	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/transport"
)

// Info describes your service
type Info struct {
	Name    string
	Actions []string // a list of actions your service wants to receive
}

// A Service manages the connection to the stark network for you.
type Service struct {
	transport.Conn
	info Info
}

// Connect creates a new service and connects it to the stark network.
func Connect(url string, info Info) (*Service, error) {
	conn, err := transport.Dial(url)
	if err != nil {
		return nil, err
	}

	return New(conn, info)
}

// MustConnect creates a new service and connects it to the stark network.
// If there is a connection error, it panics.
func MustConnect(url string, info Info) *Service {
	s, err := Connect(url, info)
	if err != nil {
		panic(err)
	}
	return s
}

// New creates a new service that listens/sends on a stark connection.
func New(conn transport.Conn, info Info) (*Service, error) {
	s := &Service{conn, info}
	if conn != nil {
		msg := stark.NewMessage()
		msg.Action = "route.hello"
		msg.Data["name"] = info.Name
		msg.Data["actions"] = info.Actions
		msg.Message = "Hello from service " + info.Name
		s.Write(msg)
	}
	return s, nil
}

// Name returns the name of this service as set in the Info.
func (s *Service) Name() string {
	return s.info.Name
}

// Write writes a message to the underlying connection and checks it for validity.
// It also correctly sets the source of the message to the name of your service.
func (s *Service) Write(msg *stark.Message) error {
	if msg.Source == "" {
		msg.Source = s.info.Name
	}
	if ok, err := msg.IsValid(); !ok {
		return err
	}
	return s.Conn.Write(msg)
}

// Read listens on the connection and returns a new message or an error.
func (s *Service) Read() (*stark.Message, error) {
	msg, err := s.Conn.Read()
	if err != nil {
		return nil, err
	}
	if msg.Action == "route.hello" {
		return s.Read()
	}
	return msg, err
}

// Handler accepts messages and optionally returns a reply.
type Handler interface {
	Handle(*stark.Message) (*stark.Message, error)
}

// HandleLoop continuously listens on the connection and calls the Handler
// if a new one was received. If the Handler returns a reply, it will be send back
// over the connection.
func (s *Service) HandleLoop(handler Handler) error {
	for {
		msg, err := s.Read()
		if err != nil {
			return err
		}
		msg, err = handler.Handle(msg)
		if err != nil {
			return err
		}
		if msg == nil {
			continue
		}
		err = s.Write(msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// HandleFunc wraps a function to expose a Handler interface.
type HandleFunc func(*stark.Message) (*stark.Message, error)

// Handle accepts messages and optionally returns a reply.
func (f HandleFunc) Handle(msg *stark.Message) (*stark.Message, error) {
	return f(msg)
}
