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

const (
	ThisAlgorithmBecomingSkynetCost = 999999999
)

func ParseSimple(text string) (proto.Message, bool) {
	msg := proto.Message{}

	// Raw JSON message
	if strings.HasPrefix(text, "{") {
		if err := json.Unmarshal([]byte(text), &msg); err == nil {
			return msg, true
		}
	}

	// "!cmd arguments", simple commands
	if strings.HasPrefix(text, "!") || strings.HasPrefix(text, ".") {
		isCmd := strings.HasPrefix(text, "!")

		text = strings.TrimLeft(text, "!. ")
		parts := strings.Split(text, " ")
		if parts[0] == "" {
			return msg, false
		}
		if isCmd {
			msg.Action = "cmd/" + parts[0]
		} else {
			msg.Action = parts[0]
		}

		msg.Text = ""
		payload := make(map[string]interface{}, 0)
		for _, part := range parts[1:] {
			keyval := strings.SplitN(part, "=", 2)
			if len(keyval) == 1 {
				if msg.Text != "" {
					msg.Text += " "
				}
				msg.Text += keyval[0]
				continue
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

	return msg, false
}

func FormatSimple(msg proto.Message) string {
	if msg.Text != "" {
		return msg.Text
	}

	return fmt.Sprintf("%s from %s.", msg.Action, msg.Source)
}
