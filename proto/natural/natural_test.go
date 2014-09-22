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
	text            string
	ok              bool
	expected        proto.Message
	expectedPayload map[string]interface{}
}

func TestParseSimple(t *testing.T) {
	tests := []simple{
		simple{"ping", true, proto.Message{
			Action: "ping",
		}, nil},

		simple{"ping device=me", true, proto.Message{
			Action:      "ping",
			Destination: "me",
		}, nil},

		simple{"ping device=me host=another", true, proto.Message{
			Action:      "ping",
			Destination: "me",
		}, map[string]interface{}{
			"host": "another",
		}},

		simple{"ping no", false, proto.Message{}, nil},
		simple{"", false, proto.Message{}, nil},
	}

	for _, test := range tests {
		msg, ok := ParseSimple(test.text)
		if ok != test.ok {
			t.Errorf("'%s' should parse? exp: %v, got: %v", test.text, test.ok, ok)
		} else if ok {
			var payload map[string]interface{}
			msg.DecodePayload(&payload)
			msg.Payload = nil
			if !reflect.DeepEqual(msg, test.expected) {
				t.Errorf("decoded message differs\nexp '%v'\ngot '%v'", test.expected, msg)
			}
			if !reflect.DeepEqual(payload, test.expectedPayload) {
				t.Errorf("decoded payload differs\nexp '%v'\ngot '%v'", test.expectedPayload, payload)
			}
		}
	}
}
