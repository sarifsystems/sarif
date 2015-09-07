// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"crypto/rand"
	"errors"
	"net/url"
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

func fromTopic(topic string) (string, string) {
	action, device := "", ""
	foundDev, foundAction := false, false
	for _, p := range topicParts(topic) {
		if p == "dev" && !foundDev {
			foundDev = true
			continue
		}
		if p == "action" && !foundAction {
			foundAction = true
			continue
		}
		if foundAction {
			action += "/" + p
		} else {
			device += "/" + p
		}
	}
	action = strings.TrimLeft(action, "/")
	device = strings.TrimLeft(device, "/")
	return action, device
}

func topicParts(action string) []string {
	if action == "" {
		return []string{}
	}
	return strings.Split(action, "/")
}

func ConvertURL(u string) (Message, error) {
	msg := Message{}
	us, err := url.Parse(u)
	if err != nil {
		return msg, err
	}
	if us.Scheme != "stark" {
		return msg, errors.New("Expected url scheme stark://")
	}

	msg.Action = us.Host + us.Path
	if us.User != nil {
		msg.Destination = us.User.Username()
	}
	err = msg.EncodePayload(us.Query())
	return msg, err
}
