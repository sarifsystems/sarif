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

func TestActionParents(t *testing.T) {
	a := ActionParents("some/action/with/sub")
	t.Log(a)
	if len(a) != 4 {
		t.Fatal("expected 4 actions, not ", len(a))
	}
	if a[0] != "some" {
		t.Error("expected some, not", a[0])
	}
	if a[1] != "some/action" {
		t.Error("expected some/action, not", a[1])
	}
	if a[2] != "some/action/with" {
		t.Error("expected some/action/with, not", a[2])
	}
	if a[3] != "some/action/with/sub" {
		t.Error("expected some/action/with/sub, not", a[3])
	}
}
