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
	msg, ok := s.Parse("Record that I acquire Nutella.")
	t.Log(msg)
	if !ok {
		t.Error("expected message to parse")
	}
	if msg.Action != "event/new" {
		t.Error("wrong action", msg.Action)
	}
	if msg.Text != "I acquire Nutella." {
		t.Error("wrong text", msg.Text)
	}
	got := struct {
		Verb   string
		Object string
	}{}
	msg.DecodePayload(&got)
	if got.Verb != "acquire" {
		t.Error("wrong verb", got.Verb)
	}
	if got.Object != "Nutella" {
		t.Error("wrong object", got.Object)
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
	got2 := struct {
		Duration string
	}{}
	msg.DecodePayload(&got2)
	if got2.Duration != "5h10m" {
		t.Error("wrong duration", got2.Duration)
	}
	if msg.Text != "stop programming" {
		t.Error("stop programming", msg.Text)
	}
}
