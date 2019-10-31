// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package amqp

import (
	"strings"
)

func getTopic(action, device string) string {
	t := ""
	if device != "" {
		t += "/dev/" + device
	}
	if action != "" {
		t += "/action/" + action
	}
	t = strings.TrimLeft(t, "/")
	t = strings.Replace(t, "/", ".", -1)
	return t
}

func fromTopic(topic string) (string, string) {
	topic = strings.Replace(topic, ".", "/", -1)
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
	return strings.Split(action, ".")
}
