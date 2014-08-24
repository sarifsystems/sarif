package mux

import (
	"strings"

	"github.com/xconstruct/stark/proto"
)

type Subscription struct {
	Action  string
	Device  string
	Handler proto.Handler
}

func (s Subscription) Matches(msg proto.Message) bool {
	if msg.Device != s.Device {
		return false
	}
	if s.Action != "" && !strings.HasPrefix(msg.Action+"/", s.Action+"/") {
		return false
	}
	return true
}

type Mux struct {
	subscriptions []Subscription
}

func New() *Mux {
	return &Mux{
		make([]Subscription, 0),
	}
}

func (m *Mux) RegisterHandler(action, device string, h proto.Handler) {
	m.subscriptions = append(m.subscriptions, Subscription{
		action,
		device,
		h,
	})
}

func (m *Mux) Handle(msg proto.Message) {
	for _, s := range m.subscriptions {
		if s.Matches(msg) {
			go s.Handler(msg)
		}
	}
}
