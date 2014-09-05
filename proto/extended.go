// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import "strings"

type Subscription struct {
	Action  string
	Device  string
	Handler Handler
}

func (s Subscription) Matches(msg Message) bool {
	if msg.Destination != s.Device {
		return false
	}
	if s.Action != "" && !strings.HasPrefix(msg.Action+"/", s.Action+"/") {
		return false
	}
	return true
}

func Subscribe(action, device string) Message {
	return Message{
		Action: "proto/sub",
		Payload: map[string]interface{}{
			"action": action,
			"device": device,
		},
	}
}

func BadRequest(reason error) Message {
	str := "Bad Request"
	if reason != nil {
		str += " - " + reason.Error()
	}
	return Message{
		Action: "nack/400",
		Payload: map[string]interface{}{
			"text": str,
		},
	}
}

func InternalError(reason error) Message {
	str := "Internal Error"
	if reason != nil {
		str += " - " + reason.Error()
	}
	return Message{
		Action: "nack/500",
		Payload: map[string]interface{}{
			"text": str,
		},
	}
}
