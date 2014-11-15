// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"testing"
)

func TestGenerateId(t *testing.T) {
	id := GenerateId()
	if len(id) != 8 {
		t.Errorf("Incorrect length of id '%s'", id)
	}
	if id == GenerateId() {
		t.Errorf("Id collision: '%s'", id)
	}
}

func TestGetTopic(t *testing.T) {
	var tp string

	tp = getTopic("", "mydevice")
	if tp != "dev/mydevice" {
		t.Errorf("Incorrect topic: %s", tp)
	}

	tp = getTopic("myaction", "")
	if tp != "action/myaction" {
		t.Errorf("Incorrect topic: %s", tp)
	}

	tp = getTopic("myaction", "mydevice")
	if tp != "dev/mydevice/action/myaction" {
		t.Errorf("Incorrect topic: %s", tp)
	}
}
