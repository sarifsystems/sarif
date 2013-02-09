package service

import (
	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/transport"
)

type Info struct {
	Name string
	Actions []string
}

type Service struct {
	stark.Conn
	info Info
}

func Connect(url string, info Info) (*Service, error) {
	conn, err := transport.Connect(url)
	if err != nil {
		return nil, err
	}

	return New(conn, info)
}

func New(conn stark.Conn, info Info) (*Service, error) {
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
