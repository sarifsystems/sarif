// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"testing"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
)

func TestParse(t *testing.T) {
	deps := &Dependencies{}
	core.InjectTest(deps)
	srv := NewService(deps)

	parsed, ok := srv.parseNatural(proto.Message{
		Action: "natural/parse",
		Text:   "When did I last visit Berlin?",
	})
	if !ok {
		t.Fatal("message should parse")
	}
	t.Log("parsed:", parsed)
	if parsed.Action != "location/last" {
		t.Error("wrong action: ", parsed.Action)
	}

	got := struct {
		Address string
	}{}
	parsed.DecodePayload(&got)
	if got.Address != "Berlin" {
		t.Error("wrong address: ", got.Address)
	}
	if parsed.Text != "When did I last visit Berlin" {
		t.Error("wrong text: ", parsed.Text)
	}
}
