// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sarif

import "time"

type subscription struct {
	Action  string        `json:"action,omitempty"`
	Device  string        `json:"device,omitempty"`
	Handler func(Message) `json:"-"`
}

func (s subscription) Matches(msg Message) bool {
	if msg.Destination != s.Device {
		return false
	}
	if !msg.IsAction(s.Action) {
		return false
	}
	return true
}

func Subscribe(action, device string) Message {
	return CreateMessage("proto/sub", subscription{action, device, nil})
}

func BadRequest(reason error) Message {
	str := "Bad Request"
	if reason != nil {
		str += " - " + reason.Error()
	}
	return Message{
		Action: "err/badrequest",
		Text:   str,
	}
}

func InternalError(reason error) Message {
	str := "Internal Error"
	if reason != nil {
		str += " - " + reason.Error()
	}
	return Message{
		Action: "err/internal",
		Text:   str,
	}
}

type ClientInfo struct {
	Name         string    `json:"name,omitempty"`
	Auth         string    `json:"auth,omitempty"`
	Capabilities []string  `json:"capabilities,omitempty"`
	LastSeen     time.Time `json:"last_seen,omitempty"`
}
