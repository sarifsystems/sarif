// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"crypto/rand"
	"strings"
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
	t := ""
	if device != "" {
		t += "/dev/" + device
	}
	if action != "" {
		t += "/action/" + action
	}
	return strings.TrimLeft(t, "/")
}

func ActionParents(action string) []string {
	parts := strings.Split(action, "/")
	pre := ""
	for i, part := range parts {
		full := pre + part
		parts[i] = full
		pre = full + "/"
	}
	return parts
}
