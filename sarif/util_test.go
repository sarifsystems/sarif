// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sarif

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
	if msg.Version == "" {
		msg.Version = VERSION
	}
	if msg.Source == "" {
		msg.Source = s.Name
	}
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

func TestFromTopic(t *testing.T) {
	a, d := "with/action", "my/device"
	a2, d2 := fromTopic(getTopic(a, d))
	if a2 != a || d2 != d {
		t.Errorf("expected %s|%s, got %s|%s", a, d, a2, d2)
	}

	a, d = "with/action", ""
	a2, d2 = fromTopic(getTopic(a, d))
	if a2 != a || d2 != d {
		t.Errorf("expected %s|%s, got %s|%s", a, d, a2, d2)
	}

	a, d = "", "my/device"
	a2, d2 = fromTopic(getTopic(a, d))
	if a2 != a || d2 != d {
		t.Errorf("expected %s|%s, got %s|%s", a, d, a2, d2)
	}

	a, d = "", ""
	a2, d2 = fromTopic(getTopic(a, d))
	if a2 != a || d2 != d {
		t.Errorf("expected %s|%s, got %s|%s", a, d, a2, d2)
	}
}
