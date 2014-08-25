// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"reflect"
	"testing"

	"github.com/xconstruct/stark/proto"
)

type simple struct {
	text     string
	ok       bool
	expected proto.Message
}

func TestParseSimple(t *testing.T) {
	tests := []simple{
		simple{"ping", true, proto.Message{
			Action: "ping",
		}},

		simple{"ping device=me", true, proto.Message{
			Action:      "ping",
			Destination: "me",
		}},

		simple{"ping device=me host=another", true, proto.Message{
			Action:      "ping",
			Destination: "me",
			Payload: map[string]interface{}{
				"host": "another",
			},
		}},

		simple{"ping no", false, proto.Message{}},
		simple{"", false, proto.Message{}},
	}

	for _, test := range tests {
		msg, ok := ParseSimple(test.text)
		if ok != test.ok {
			t.Errorf("'%s' should parse? exp: %v, got: %v", test.text, test.ok, ok)
		} else if ok && !reflect.DeepEqual(msg, test.expected) {
			t.Errorf("decoded message differs\nexp '%v'\ngot '%v'", test.expected, msg)
		}
	}
}
