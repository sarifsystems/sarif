package natural

import (
	"strings"
	"github.com/xconstruct/stark/proto"
)

var actions = map[string]string{
	"play": "play",
	"pause": "pause",
	"stop": "stop",
	"next": "next",
	"prev": "prev",
}

var destinations = map[string]string{
	"mpd": "mpd",
	"music": "mpd",
	"song": "mpd",
}

func Parse(text string) *proto.Message {
	text = strings.TrimSpace(text)
	words := strings.Split(text, " ")

	var action, dest string
	for _, word := range words {
		if word == "echo" {
			action = "echo"
			dest = "echo"
			break
		}
		if actions[word] != "" {
			action = actions[word]
			continue
		}
		if destinations[word] != "" {
			dest = destinations[word]
			continue
		}
	}

	if action == "" || dest == "" {
		return nil
	}

	msg := proto.NewMessage()
	msg.Action = action
	msg.Destination = dest
	msg.Message = text
	return msg
}

const ACTION_PROCESS string = "process_natural"

func Handle (msg *proto.Message) *proto.Message {
	if msg.Action != ACTION_PROCESS {
		return nil
	}

	reply := Parse(msg.Message)
	if reply == nil {
		reply = proto.NewMessage()
		reply.Action = "error"
		reply.Source = "natural"
		reply.Destination = msg.Source
		reply.Message = "Did not understand: " + msg.Message
		return reply
	}
	reply.Source = msg.Source // TODO: Reply to?
	return reply
}
