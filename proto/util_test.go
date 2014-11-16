// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import "testing"

type testService struct {
	Name     string
	T        *testing.T
	Conn     Conn
	Received []Message
}

func newTestService(name string, t *testing.T) *testService {
	return &testService{
		name,
		t,
		nil,
		make([]Message, 0),
	}
}

func (s *testService) Listen(conn Conn) {
	s.Conn = conn
	for {
		msg, err := conn.Read()
		if err != nil {
			s.T.Fatal(err)
		}
		s.Received = append(s.Received, msg)
	}
}

func (s *testService) NewLocalConn() Conn {
	a, b := NewPipe()
	go s.Listen(a)
	return b
}

func (s *testService) Publish(msg Message) {
	if err := s.Conn.Write(msg); err != nil {
		s.T.Fatal(err)
	}
}

func (s *testService) Reset() {
	s.Received = make([]Message, 0)
}

func (s *testService) Fired() bool {
	return len(s.Received) > 0
}
