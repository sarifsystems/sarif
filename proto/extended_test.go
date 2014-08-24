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
