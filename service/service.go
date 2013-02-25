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

// Handler accepts messages and optionally returns a reply.
type Handler interface {
	Handle(*stark.Message) (*stark.Message, error)
}

// Info describes your service
type Info struct {
	Name    string
	Actions []string // a list of actions your service wants to receive
}

// A Service manages the connection to the stark network for you.
type Service struct {
	transport.Conn
	info Info
	Handler Handler
}

// New creates a new service that listens/sends on a stark connection.
func New(info Info) *Service {
	return &Service{info: info}
}

// Name returns the name of this service as set in the Info.
func (s *Service) Name() string {
	return s.info.Name
}

func (s *Service) Dial(url string) error {
	conn, err := transport.Dial(url)
	if err != nil {
		return err
	}
	return s.Connect(conn)
}

func (s *Service) Connect(conn transport.Conn) error {
	s.Conn = conn
	msg := stark.NewMessage()
	msg.Action = "route.hello"
	msg.Data["name"] = s.info.Name
	msg.Data["actions"] = s.info.Actions
	msg.Message = "Hello from service " + s.info.Name
	return s.Write(msg)
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

// Serve continuously listens on the connection and calls the Handler
// if a new one was received. If the Handler returns a reply, it will be send back
// over the connection.
func (s *Service) Serve() error {
	for {
		msg, err := s.Read()
		if err != nil {
			return err
		}
		msg, err = s.Handler.Handle(msg)
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
