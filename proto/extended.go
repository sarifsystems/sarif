// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

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
	return Message{
		Action: "nack/400",
		Payload: map[string]interface{}{
			"text": reason,
		},
	}
}
