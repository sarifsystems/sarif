package proto

import (
	"encoding/json"
)

type ByteReadWriter interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
}

type ByteEndpoint struct {
	enc     *json.Encoder
	dec     *json.Decoder
	handler Handler
}

func NewByteEndpoint(conn ByteReadWriter) *ByteEndpoint {
	t := &ByteEndpoint{
		json.NewEncoder(conn),
		json.NewDecoder(conn),
		nil,
	}
	return t
}

func (t *ByteEndpoint) Publish(msg Message) error {
	return t.enc.Encode(msg)
}

func (t *ByteEndpoint) RegisterHandler(h Handler) {
	t.handler = h
}

func (t *ByteEndpoint) Listen() error {
	for {
		var msg Message
		if err := t.dec.Decode(&msg); err != nil {
			return err
		}

		if t.handler != nil {
			t.handler(msg)
		}
	}
}
