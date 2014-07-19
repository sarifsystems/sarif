package client

import (
	"reflect"
	"testing"
)

func TestMessageEncoding(t *testing.T) {
	m := Message{
		Version: VERSION,
		Id:      GenerateId(),
		Action:  "testaction",
		Source:  "testsource",
	}
	t.Log(m)

	enc, err := m.Encode()
	if err != nil {
		t.Error(err)
	}
	t.Log(string(enc))

	dec, err := DecodeMessage(enc)
	t.Log(dec)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(dec, m) {
		t.Error("decoded message differs")
	}
}

func TestValid(t *testing.T) {
	var m Message
	m = Message{Version: "0.2", Id: "12345678", Action: "testaction", Source: "testsource"}
	if err := m.IsValid(); err != nil {
		t.Error(err)
	}

	m = Message{Version: "0.2", Id: "", Action: "testaction", Source: "testsource"}
	if err := m.IsValid(); err == nil {
		t.Error("Message without id passes as valid")
	}

	m = Message{Version: "", Id: "12345678", Action: "testaction", Source: "testsource"}
	if err := m.IsValid(); err == nil {
		t.Error("Message without version passes as valid")
	}

	m = Message{Version: "0.2", Id: "12345678", Action: "", Source: "testsource"}
	if err := m.IsValid(); err == nil {
		t.Error("Message without action passes as valid")
	}

	m = Message{Version: "", Id: "12345678", Action: "testaction", Source: ""}
	if err := m.IsValid(); err == nil {
		t.Error("Message without source passes as valid")
	}
}

func TestReply(t *testing.T) {
	orig := Message{Version: VERSION, Id: GenerateId(), Action: "ping", Source: "originaldevice"}
	reply := orig.Reply(Message{Version: VERSION, Id: GenerateId(), Action: "ack", Source: "newdevice"})

	if reply.Id == orig.Id {
		t.Error("Reply has same id:", reply.Id)
	}
	if reply.CorrId != orig.Id {
		t.Error("Reply has wrong corrId:", reply.CorrId)
	}
	if reply.Device != orig.Source {
		t.Error("Reply has wrong device:", reply.Device)
	}
}
