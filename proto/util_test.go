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

	tp = GetTopic("", "mydevice")
	if tp != "stark/dev/mydevice" {
		t.Errorf("Incorrect topic: %s", tp)
	}

	tp = GetTopic("myaction", "")
	if tp != "stark/special/all/action/myaction" {
		t.Errorf("Incorrect topic: %s", tp)
	}

	tp = GetTopic("myaction", "mydevice")
	if tp != "stark/dev/mydevice/action/myaction" {
		t.Errorf("Incorrect topic: %s", tp)
	}
}
