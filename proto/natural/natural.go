// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xconstruct/stark/proto"
)

func ParseSimple(text string) (proto.Message, bool) {
	msg := proto.Message{}

	if strings.HasPrefix(text, "{") {
		if err := json.Unmarshal([]byte(text), &msg); err == nil {
			return msg, true
		}
	}

	parts := strings.Split(text, " ")
	msg.Action = parts[0]
	if msg.Action == "" {
		return msg, false
	}

	payload := make(map[string]interface{}, 0)
	for _, part := range parts[1:] {
		keyval := strings.SplitN(part, "=", 2)
		if len(keyval) == 1 {
			return msg, false
		}

		k, v := keyval[0], keyval[1]
		switch k {
		case "text":
			msg.Text = v
		case "device":
			fallthrough
		case "destination":
			msg.Destination = v
		default:
			payload[k] = v
		}
	}
	if len(payload) > 0 {
		msg.EncodePayload(payload)
	}
	return msg, true
}

func FormatSimple(msg proto.Message) string {
	if msg.Text != "" {
		return msg.Text
	}

	return fmt.Sprintf("%s from %s.", msg.Action, msg.Source)
}
