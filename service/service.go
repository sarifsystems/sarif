package service

import (
	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/transport"
)

type Handler interface {
	Handle(*stark.Message) (*stark.Message, error)
}

type Info struct {
	Name string
	Actions []string
}

type Service struct {
	transport.Conn
	info Info
}

func Connect(url string, info Info) (*Service, error) {
	conn, err := transport.Connect(url)
	if err != nil {
		return nil, err
	}

	return New(conn, info)
}

func MustConnect(url string, info Info) *Service {
	s, err := Connect(url, info)
	if err != nil {
		panic(err)
	}
	return s
}

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

func (s *Service) Name() string {
	return s.info.Name
}

func (s *Service) Write(msg *stark.Message) error {
	if msg.Source == "" {
		msg.Source = s.info.Name
	}
	if ok, err := msg.IsValid(); !ok {
		return err
	}
	return s.Conn.Write(msg)
}

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

func Matches(action, pattern string) bool {
	n := len(action)
	return len(pattern) >= n && pattern[0:n] == action
}

type HandleFunc func(*stark.Message) (*stark.Message, error)

func (f HandleFunc) Handle(msg *stark.Message) (*stark.Message, error) {
	return f(msg)
}
