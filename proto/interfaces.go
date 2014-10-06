// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

type Handler func(msg Message)
type Publisher func(msg Message) error

type Conn interface {
	Publish(msg Message) error
	RegisterHandler(h Handler)
}

type Service interface {
	Handle(msg Message)
	RegisterPublisher(p Publisher)
}

func Connect(c Conn, s Service) {
	s.RegisterPublisher(c.Publish)
	c.RegisterHandler(s.Handle)
}
