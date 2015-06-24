// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"testing"

	"github.com/xconstruct/stark/core"
)

func TestParse(t *testing.T) {
	deps := &Dependencies{}
	core.InjectTest(deps)
	srv := NewService(deps)
	if err := srv.Enable(); err != nil {
		t.Fatal(err)
	}

	parsed, err := srv.parser.Parse("When did I last visit Berlin?", &Context{})
	if err != nil {
		t.Fatal("message should parse")
	}
	t.Log("parsed:", parsed)
	if parsed.Message.Action != "location/last" {
		t.Error("wrong action: ", parsed.Message.Action)
	}

	got := struct {
		Address string
	}{}
	parsed.Message.DecodePayload(&got)
	if got.Address != "Berlin" {
		t.Error("wrong address: ", got.Address)
	}
	if parsed.Text != "When did I last visit Berlin?" {
		t.Error("wrong text: ", parsed.Message.Text)
	}
}
