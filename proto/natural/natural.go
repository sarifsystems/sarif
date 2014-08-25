package natural

import (
	"fmt"
	"strings"

	"github.com/xconstruct/stark/proto"
)

func ParseSimple(text string) (proto.Message, bool) {
	msg := proto.Message{}
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
		case "device":
			fallthrough
		case "destination":
			msg.Destination = v
		default:
			payload[k] = v
		}
	}
	if len(payload) > 0 {
		msg.Payload = payload
	}
	return msg, true
}

func FormatSimple(msg proto.Message) string {
	if text := msg.PayloadGetString("text"); text != "" {
		return text
	}

	return fmt.Sprintf("%s from %s.", msg.Action, msg.Source)
}
