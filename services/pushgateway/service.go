// Copyright (C) 2017 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service pushgateway provides notifications to FCM
package pushgateway

import (
	"errors"
	"fmt"
	"strings"

	"github.com/maddevsio/fcm"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
)

var Module = &services.Module{
	Name:        "pushgateway",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	Config services.Config
	Client *sarif.Client
}

type Config struct {
	FirebaseToken string

	Clients map[string]string
}

type client struct {
	Name  string
	Token string
	Queue []sarif.Message
}

type Service struct {
	Config Config
	cfg    services.Config
	*sarif.Client
	fcm     *fcm.FCM
	clients map[string]*client
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Client: deps.Client,
		cfg:    deps.Config,
	}
	return s
}

func (s *Service) Enable() (err error) {
	s.Config.Clients = make(map[string]string)
	s.cfg.Get(&s.Config)
	s.fcm = fcm.NewFCM(s.Config.FirebaseToken)
	s.clients = make(map[string]*client, 0)

	s.Subscribe("push/register", "", s.handlePushRegister)
	s.Subscribe("push/fetch", "", s.handlePushFetch)
	for client, token := range s.Config.Clients {
		s.initClient(client, token)
	}
	return nil
}

type RegisterPayload struct {
	Name  string `json:"name,omitempty"`
	Token string `json:"token"`
}

func (s *Service) handlePushRegister(msg sarif.Message) {
	var p RegisterPayload
	if err := msg.DecodePayload(&p); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	if p.Token == "" {
		s.ReplyBadRequest(msg, errors.New("No token given"))
		return
	}
	if p.Name == "" {
		p.Name = msg.Source
	}

	s.Config.Clients[p.Name] = p.Token
	s.initClient(p.Name, p.Token)
	s.cfg.Set(s.Config)
	s.Reply(msg, sarif.CreateMessage("push/registered", nil))
}

func (s *Service) handleIncoming(msg sarif.Message) {
	name := strings.TrimPrefix(msg.Destination, "push/")
	client := s.clients[name]
	if client == nil {
		s.Log("err/internal", "unknown client '"+name+"'")
		return
	}

	client.Queue = append(client.Queue, msg)
	if len(client.Queue) > 30 {
		client.Queue = client.Queue[1:]
	}

	data := map[string]string{
		"id":     msg.Id,
		"action": msg.Action,
	}
	_, err := s.fcm.Send(fcm.Message{
		Data:             data,
		RegistrationIDs:  []string{client.Token},
		ContentAvailable: true,
		Priority:         fcm.PriorityHigh,
	})
	if err != nil {
		s.Log("err/internal", "fcm error: "+err.Error())
	}
}

func (s *Service) initClient(name, token string) {
	if client := s.clients[name]; client != nil {
		client.Token = token
		return
	}

	s.clients[name] = &client{
		Name:  name,
		Token: token,
		Queue: make([]sarif.Message, 0),
	}
	s.Subscribe("", "push/"+name, s.handleIncoming)
}

func (s *Service) handlePushFetch(msg sarif.Message) {
	name := strings.TrimPrefix(msg.Destination, "push/fetch/")
	if name == "" {
		name = msg.Source
	}

	client := s.clients[name]
	if client == nil {
		s.ReplyBadRequest(msg, fmt.Errorf("unknown client %q", name))
		return
	}

	s.Reply(msg, sarif.CreateMessage("push/fetched", map[string]int{
		"num_messages": len(client.Queue),
	}))

	for _, queued := range client.Queue {
		if queued.CorrId == "" {
			queued.CorrId = queued.Id
		}
		queued.Id = sarif.GenerateId()
		queued.Destination = msg.Source
		s.Publish(queued)
	}

	client.Queue = make([]sarif.Message, 0)
}
