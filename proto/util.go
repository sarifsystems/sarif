// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"crypto/rand"
)

func GenerateId() string {
	const alphanum = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	var bytes = make([]byte, 8)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func getTopic(action, device string) string {
	t := "stark"
	if device != "" {
		t += "/dev/" + device
	} else {
		t += "/special/all"
	}
	if action != "" {
		t += "/action/" + action
	}
	return t
}
