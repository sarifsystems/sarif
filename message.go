// Package stark provides the core message protocol of stark.
//
// It also includes encoding/decoding to json as well as helper functions.
package stark

import (
	"crypto/rand"
	"fmt"
	"encoding/json"
)

// The supported protocol version.
const VERSION string = "0.1"

// The Go-representation of a stark message.
// For more information, please see the protocol spec.
type Message struct {
	Version string `json:"version"`
	UUID string `json:"uuid"`
	Action string `json:"action"`
	Source string `json:"source"`

	Destination string `json:"destination,omitempty"`
	ReplyTo string `json:"reply_to,omitempty"`

	Data map[string]interface{} `json:"data,omitempty"`

	Cause string `json:"cause,omitempty"`
	CausedBy string `json:"caused_by,omitempty"`
	Message string `json:"message`
}

// String returns a simple human-readable form of the message.
func (m *Message) String() string {
	path := m.Source
	if path == "" {
		path = "unknown"
	}
	if m.Destination != "" {
		path += " -> " + m.Destination
	}
	return fmt.Sprintf(`stark.Message(%s: %s "%s")`,
		path, m.Action, m.Message,
	)
}

// Decode converts a JSON message into a Message struct.
func Decode(msg []byte) (*Message, error) {
	var m Message
	err := json.Unmarshal(msg, &m)
	if valid, err := m.IsValid(); !valid {
		return &m, err
	}

	return &m, err
}

// Encode converts a Message struct into the equivalent JSON message.
func Encode(m *Message) ([]byte, error) {
	if valid, err := m.IsValid(); !valid {
		return nil, err
	}

	data, err := json.Marshal(m)
	return data, err
}

// GenerateUUID generates a unique ID for message identification.
func GenerateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// NewMessage returns new Message filled with the current version and a new UUID.
func NewMessage() (*Message) {
	m := &Message{}
	m.Version = VERSION
	m.UUID = GenerateUUID()
	m.Data = make(map[string]interface{})
	return m
}

func NewMessageFromTemplate(msg *Message) (*Message) {
	m := &(*msg)
	m.Version = VERSION
	m.UUID = GenerateUUID()
	return m
}

// This error describes a Message which is not valid as described in the spec.
type InvalidMessageError struct {
	S string
}

func (e *InvalidMessageError) Error() string {
	return "proto: " + e.S
}

// IsValid returns true if the message complies with the spec.
// Otherwise a error describing the violation is returned.
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
	if len(m.Source) == 0 {
		return false, &InvalidMessageError{"No source specified"}
	}

	return true, nil
}

// NewReply generates a new reply to the passed Message.
// The Destination is the source of the original message or, if set, the
// target specified in the ReplyTo field.
func NewReply(m *Message) *Message {
	reply := NewMessage()
	reply.Destination = m.Source
	if m.ReplyTo != "" {
		reply.Destination = m.ReplyTo
	}
	reply.Cause = "reply"
	reply.CausedBy = m.UUID
	if m.CausedBy != "" {
		reply.CausedBy = m.CausedBy
	}
	return reply
}
