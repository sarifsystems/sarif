// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import "testing"

func TestParseRegular(t *testing.T) {
	s, err := LoadRegularSchemata(defaultRegularText)
	if err != nil {
		t.Fatal(err)
	}

	// the nutella test
	msg, ok := s.Parse("I acquire Nutella.")
	t.Log(msg)
	if !ok {
		t.Error("expected message to parse")
	}
	if msg.Action != "event/new" {
		t.Error("wrong action", msg.Action)
	}
	if v := msg.PayloadGetString("verb"); v != "acquire" {
		t.Error("wrong verb", v)
	}
	if v := msg.PayloadGetString("object"); v != "Nutella" {
		t.Error("wrong object", v)
	}

	// more complicted
	msg, ok = s.Parse("remind me in 5h10m to stop programming")
	t.Log(msg)
	if !ok {
		t.Error("expected message to parse")
	}
	if msg.Action != "schedule/duration" {
		t.Error("wrong action", msg.Action)
	}
	if v := msg.PayloadGetString("duration"); v != "5h10m" {
		t.Error("wrong duration", v)
	}
	if v := msg.PayloadGetString("text"); v != "stop programming" {
		t.Error("stop programming", v)
	}
}
