package service

import (
	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/transport"
)

type Service struct {
	name string
	stark.Conn
}

func Connect(name, url string) (*Service, error) {
	conn, err := transport.Connect(url)
	if err != nil {
		return nil, err
	}

	return New(name, conn)
}

func New(name string, conn stark.Conn) (*Service, error) {
	s := &Service{name, conn}
	return s, nil
}

func (s *Service) Name() string {
	return s.name
}

func (s *Service) Write(msg *stark.Message) error {
	if msg.Source == "" {
		msg.Source = s.name
	}
	return s.Conn.Write(msg)
}
