// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sarif

import "time"

type Client interface {
	DeviceId() string
	Publish(msg Message) error
	Subscribe(action, device string, h func(Message)) error

	Request(msg Message) <-chan Message
	Reply(orig, reply Message) error
	ReplyBadRequest(orig Message, err error) error
	ReplyInternalError(orig Message, err error) error

	Discover(action string) <-chan Message
	Log(typ, text string, args ...interface{}) error
}

type ClientInfo struct {
	Name         string    `json:"name,omitempty"`
	Auth         string    `json:"auth,omitempty"`
	Capabilities []string  `json:"capabilities,omitempty"`
	LastSeen     time.Time `json:"last_seen,omitempty"`
}
