// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service states tracks current state.
package states

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "states",
	Version:     "1.0",
	NewInstance: NewService,
}

type State struct {
	Value    interface{}
	Modified time.Time
}

type Dependencies struct {
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	Log proto.Logger
	*proto.Client

	States map[string]State
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Log:    deps.Log,
		Client: deps.Client,
		States: make(map[string]State),
	}
	return s
}

func (s *Service) Enable() error {
	s.Subscribe("state", "", s.handleChange)
	s.Subscribe("states/get", "", s.handleStatesGet)
	return nil
}

func splitValue(state string) (string, float64, bool) {
	state = strings.TrimPrefix(state, "state/")
	i := strings.LastIndex(state, "/")
	if i == -1 {
		return state, 0, false
	}

	f, err := strconv.ParseFloat(state[i+1:], 64)
	return state[0:i], f, err == nil
}

func (s *Service) handleChange(msg proto.Message) {
	state, val, ok := splitValue(msg.Action)
	if !ok {
		return
	}
	st := s.States[state]
	st.Value = val
	st.Modified = time.Now()
	s.States[state] = st
}

func (s *Service) handleStatesGet(msg proto.Message) {
	subcat := ""
	if strings.HasPrefix(msg.Action, "states/get/") {
		subcat = strings.TrimPrefix(msg.Action, "states/get/")
	}

	filtered := make(map[string]interface{})
	for name, state := range s.States {
		if name == subcat {
			filtered["value"] = state.Value
			filtered[name] = state.Value
			continue
		}
		if subcat == "" || strings.HasPrefix(name+"/", subcat+"/") {
			filtered[name] = state.Value
		}
	}
	fmt.Println(subcat, s.States, filtered)
	s.Reply(msg, proto.CreateMessage("states/got", filtered))
}
