// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"reflect"
	"testing"

	"github.com/sarifsystems/sarif/sarif"
)

type simple struct {
	text            string
	ok              bool
	expected        sarif.Message
	expectedPayload map[string]interface{}
}

func TestParseSimple(t *testing.T) {
	tests := []simple{
		simple{".ping", true, sarif.Message{
			Action: "ping",
		}, nil},

		simple{".ping device=me", true, sarif.Message{
			Action:      "ping",
			Destination: "me",
		}, nil},

		simple{".ping some text", true, sarif.Message{
			Action: "ping",
			Text:   "some text",
		}, nil},

		simple{"/ping some text", true, sarif.Message{
			Action: "ping",
			Text:   "some text",
		}, nil},

		simple{".ping with some device=me host=another things", true, sarif.Message{
			Action:      "ping",
			Destination: "me",
			Text:        "with some things",
		}, map[string]interface{}{
			"host": "another",
		}},

		simple{`.ping with "some device=me" host="another things" this`, true, sarif.Message{
			Action: "ping",
			Text:   `with some device=me this`,
		}, map[string]interface{}{
			"host": "another things",
		}},

		simple{"ping no", false, sarif.Message{}, nil},
		simple{"", false, sarif.Message{}, nil},
	}

	for _, test := range tests {
		msg, ok := ParseSimple(test.text)
		if ok != test.ok {
			t.Errorf("'%s' should parse? exp: %v, got: %v", test.text, test.ok, ok)
		} else if ok {
			var payload map[string]interface{}
			msg.DecodePayload(&payload)
			msg.Payload.Raw = nil
			if !reflect.DeepEqual(msg, test.expected) {
				t.Errorf("decoded message differs\nexp '%v'\ngot '%v'", test.expected, msg)
			}
			if !reflect.DeepEqual(payload, test.expectedPayload) {
				t.Errorf("decoded payload differs\nexp '%v'\ngot '%v'", test.expectedPayload, payload)
			}
		}
	}
}

func TestFormatMessage(t *testing.T) {
	msg := sarif.Message{
		Text: "Hello, the time is 2015-03-14T18:48:10+02:00.",
	}
	FormatMessage(&msg)
	exp := "Hello, the time is Sat, 14 Mar 15 at 18:48."
	if msg.Text != exp {
		t.Error("Unexpected format: ", msg.Text)
	}
}
