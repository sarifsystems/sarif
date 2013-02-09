package natural

import (
	"strings"
	"github.com/xconstruct/stark"
)

var actions = map[string]string{
	"mpd": "music",
	"music": "music",
	"song": "music",
}

var subactions = map[string]string{
	"play": "play",
	"pause": "pause",
	"stop": "stop",
	"next": "next",
	"prev": "prev",
}


func Parse(text string) *stark.Message {
	text = strings.TrimSpace(text)
	words := strings.Split(text, " ")

	var action, subaction string
	for _, word := range words {
		if actions[word] != "" {
			action = actions[word]
			continue
		}
		if subactions[word] != "" {
			subaction = subactions[word]
			continue
		}
	}

	if action == "" || subaction == "" {
		return nil
	}

	msg := stark.NewMessage()
	msg.Action = action + "." + subaction
	msg.Message = text
	return msg
}
