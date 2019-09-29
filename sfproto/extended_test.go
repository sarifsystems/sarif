// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sfproto

import (
	"testing"
)

func TestSubscribe(t *testing.T) {
	msg := Subscribe("act", "dev")
	var got subscription
	msg.DecodePayload(&got)
	if got.Action != "act" {
		t.Errorf("Message payload action wrong, got '%v'", got.Action)
	}
	if got.Device != "dev" {
		t.Errorf("Message payload device wrong, got '%v'", got.Device)
	}
}
