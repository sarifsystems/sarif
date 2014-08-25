package proto

import (
	"encoding/json"
	"errors"
)

const VERSION = "0.3"

type Message struct {
	Version     string      `json:"v"`
	Id          string      `json:"id"`
	Action      string      `json:"action"`
	Source      string      `json:"src"`
	Destination string      `json:"dst,omitempty"`
	Payload     interface{} `json:"p,omitempty"`
	CorrId      string      `json:"corr,omitempty"`
}

func (m Message) Encode() ([]byte, error) {
	if err := m.IsValid(); err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

func DecodeMessage(raw []byte) (Message, error) {
	m := Message{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return m, err
	}
	if err := m.IsValid(); err != nil {
		return m, err
	}
	return m, nil
}

func (m Message) IsValid() error {
	if m.Version == "" {
		return errors.New("Invalid stark message: missing version")
	}
	if m.Id == "" {
		return errors.New("Invalid stark message: missing id")
	}
	if m.Action == "" {
		return errors.New("Invalid stark message: missing action")
	}
	if m.Source == "" {
		return errors.New("Invalid stark message: missing source")
	}
	return nil
}

func (orig Message) Reply(m Message) Message {
	if m.CorrId == "" {
		m.CorrId = orig.Id
	}
	if m.Destination == "" {
		m.Destination = orig.Source
	}
	return m
}

func (m Message) PayloadGetString(key string) string {
	if pmap, ok := m.Payload.(map[string]interface{}); ok {
		if val, ok := pmap[key]; ok {
			if str, ok := val.(string); ok {
				return str
			}
		}
	}
	return ""
}
