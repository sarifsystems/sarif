// Package service provides a wrapper class to create services without the need
// to care about the underlying connection.
//
// The Service class handles such things as connecting and identifying to the router and
// broadcasting the capabilities for you.
package service

import (
	"log"
	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/transport"
)

// Info describes a service
type Info struct {
	Name    string
	Actions []string // a list of actions the service understands
}

// A Service manages the connection to the stark network for you.
type Service struct {
	Conns map[stark.Conn]*Info
	info Info
	Handler stark.Handler
}

// New creates a new service that listens/sends on a stark connection.
func New(info Info) *Service {
	return &Service{
		Conns: make(map[stark.Conn]*Info),
		info: info,
	}
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
	s.Connect(conn)
	return nil
}

func (s *Service) Listen(url string) error {
	listener, err := transport.Listen(url)
	if err != nil {
		return err
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			s.Connect(conn)
		}
	}()
	return nil
}

func (s *Service) Connect(conn stark.Conn) {
	go func() {
		s.Conns[conn] = &Info{}
		msg := stark.NewMessage()
		msg.Source = s.info.Name
		msg.Action = "route.hello"
		msg.Data["actions"] = s.info.Actions
		msg.Message = "Hello from service " + s.info.Name
		if err := conn.Write(msg); err != nil {
			s.Disconnect(conn, err)
			return
		}

		for {
			msg, err := conn.Read()
			if err != nil {
				s.Disconnect(conn, err)
				return
			}

			if msg.Action == "route.hello" {
				s.Conns[conn].Name = stark.ParsePath(msg.Source).First()
				s.Conns[conn].Actions, _ = msg.Data["actions"].([]string)
			}

			if reply := s.Handler.Handle(msg); reply != nil {
				if err := s.Write(reply); err != nil {
					s.Disconnect(conn, err)
					return
				}
			}
		}
	}()
}

func (s *Service) Disconnect(conn stark.Conn, reason error) {
	log.Printf("service/disconnect: %v\n", reason)
	delete(s.Conns, conn)
}

// Write writes a message to the underlying connection and checks it for validity.
// It also correctly sets the source of the message to the name of your service.
func (s *Service) WriteConn(conn stark.Conn, msg *stark.Message) error {
	if msg.Source == "" {
		msg.Source = s.info.Name
	}
	if ok, err := msg.IsValid(); !ok {
		return err
	}
	err := conn.Write(msg)
	if err != nil {
		s.Disconnect(conn, err)
	}
	return err
}

func (s *Service) Write(msg *stark.Message) error {
	route := stark.ParseRoute(msg.Source, msg.Destination)
	next := route.Forward(s.info.Name)
	msg.Source, msg.Destination = route.Strings()

	// Only one connection? Send to this
	if len(s.Conns) == 1 {
		for conn, _ := range s.Conns {
			return s.WriteConn(conn, msg)
		}
	}

	// Exact destination found
	if next != "" {
		for conn, info := range s.Conns {
			if info.Name != next {
				continue
			}
			return s.WriteConn(conn, msg)
		}
	}

	// Action-based routing
	for conn, info := range s.Conns {
		if info.Actions == nil {
			continue
		}
		for _, action := range info.Actions {
			if action == msg.Action {
				return s.WriteConn(conn, msg)
			}
		}
	}

	return nil
}

var ForwardHandler stark.HandleFunc = func(msg *stark.Message) *stark.Message {
	return msg
}

func LogHandler(h stark.Handler) stark.HandleFunc {
	return stark.HandleFunc(func(msg *stark.Message) *stark.Message {
		log.Println(msg)
		return h.Handle(msg)
	})
}
