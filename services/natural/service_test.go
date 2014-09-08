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
	ctx, _ := core.NewTestContext()
	srv, err := NewService(ctx)
	if err != nil {
		t.Fatal(err)
	}

	parsed, ok := srv.parseNatural(proto.Message{
		Action: "natural/parse",
		Payload: map[string]interface{}{
			"text": "When did I last visit Berlin?",
		},
	})
	if !ok {
		t.Fatal("message should parse")
	}
	t.Log("parsed:", parsed)
	if parsed.Action != "location/last" {
		t.Error("wrong action: ", parsed.Action)
	}
	if v := parsed.PayloadGetString("address"); v != "Berlin" {
		t.Error("wrong address: ", v)
	}
	if v := parsed.PayloadGetString("text"); v != "When did I last visit Berlin?" {
		t.Error("wrong text: ", v)
	}
}
