package stark

import (
	"testing"
)

func TestMinimum(t *testing.T) {
	// Specify message
	m := NewMessage()
	m.Action = "push"
	m.Source = "desktop"
	m.Destination = "sgs2"
	m.Data["url"] = "http://google.com"
	m.Message = "Push link http://google.com to sgs2"

	// Test encoding
	b, err := Encode(m)
	if err != nil {
		t.Fatalf("encoding error: %v", err)
	}

	// Test decoding
	dec, err := Decode(b)
	if err != nil {
		t.Fatalf("decoding error: %v", err)
	}
	if dec.Action != "push" {
		t.Errorf("wrong action: expected %v, got %v", "push", dec.Action)
	}

	if dec.Data["url"] != "http://google.com" {
		t.Errorf("wrong data-url: expected %v, got %v", "http://google.com", dec.Data["url"])
	}
}
