package natural

import (
	"strings"
	"github.com/xconstruct/stark"
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

func Parse(text string) *stark.Message {
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

	msg := stark.NewMessage()
	msg.Action = action
	msg.Destination = dest
	msg.Message = text
	return msg
}
