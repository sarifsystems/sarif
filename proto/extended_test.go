// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"testing"
)

func TestSubscribe(t *testing.T) {
	msg := Subscribe("act", "dev")
	if v := msg.PayloadGetString("action"); v != "act" {
		t.Errorf("Message payload action wrong, got '%v'", v)
	}
	if v := msg.PayloadGetString("device"); v != "dev" {
		t.Errorf("Message payload device wrong, got '%v'", v)
	}
}
