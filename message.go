package stark

import (
	"crypto/rand"
	"fmt"
	"encoding/json"
)

const VERSION string = "0.1"

type Message struct {
	Version string `json:"version"`
	UUID string `json:"uuid"`
	Action string `json:"action"`
	Source string `json:"source"`
	Destination string `json:"destination"`
	ReplyTo string `json:"reply_to"`

	Data map[string]interface{} `json:"data,omitempty"`

	Cause string `json:"cause,omitempty"`
	CausedBy string `json:"caused_by,omitempty"`
	Message string `json:"message`
}

func (m *Message) String() string {
	return fmt.Sprintf(`stark.Message(%s > %s: %s "%s")`,
		m.Source, m.Destination, m.Action, m.Message,
	)
}

func Decode(msg []byte) (*Message, error) {
	var m Message
	err := json.Unmarshal(msg, &m)
	if valid, err := m.IsValid(); !valid {
		return nil, err
	}

	return &m, err
}

func Encode(m *Message) ([]byte, error) {
	if valid, err := m.IsValid(); !valid {
		return nil, err
	}

	data, err := json.Marshal(m)
	return data, err
}

func GenerateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func NewMessage() (*Message) {
	m := &Message{}
	m.Version = VERSION
	m.UUID = GenerateUUID()
	m.Data = make(map[string]interface{})
	return m
}

type InvalidMessageError struct {
	S string
}

func (e *InvalidMessageError) Error() string {
	return "proto: " + e.S
}

func (m *Message) IsValid() (bool, error) {
	if m.Version != VERSION {
		return false, &InvalidMessageError{"Unsupported version: " + m.Version}
	}
	if len(m.Message) == 0 {
		return false, &InvalidMessageError{"No message specified"}
	}
	if len(m.UUID) == 0 {
		return false, &InvalidMessageError{"No UUID specified"}
	}

	return true, nil
}

func NewReply(m *Message) *Message {
	reply := NewMessage()
	reply.Source = m.Destination
	reply.Destination = m.Source
	if m.ReplyTo != "" {
		reply.Destination = m.ReplyTo
	}
	reply.Cause = "reply"
	reply.CausedBy = m.UUID
	return reply
}
