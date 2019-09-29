// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sfproto

import (
	"github.com/sarifsystems/sarif/sarif"
)

type subscription struct {
	Action  string              `json:"action,omitempty"`
	Device  string              `json:"device,omitempty"`
	Handler func(sarif.Message) `json:"-"`
}

func (s subscription) Matches(msg sarif.Message) bool {
	if msg.Destination != s.Device {
		return false
	}
	if !msg.IsAction(s.Action) {
		return false
	}
	return true
}

func Subscribe(action, device string) sarif.Message {
	return sarif.CreateMessage("proto/sub", subscription{action, device, nil})
}
