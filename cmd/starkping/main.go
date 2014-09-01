// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
)

type PingService struct {
	pings map[string]time.Time
	proto *proto.Client
}

func NewPingService(ep proto.Endpoint) *PingService {
	return &PingService{
		make(map[string]time.Time),
		proto.NewClient("starkping", ep),
	}
}

func (s *PingService) Enable() error {
	s.proto.RegisterHandler(s.Handle)
	return s.proto.SubscribeSelf("ack")
}

func (s *PingService) Handle(msg proto.Message) {
	if msg.Action != "ack" {
		return
	}
	sent, ok := s.pings[msg.CorrId]
	if !ok {
		return
	}

	fmt.Printf("%s from %s: time=%.1fms\n",
		msg.Action,
		msg.Source,
		time.Since(sent).Seconds()*1e3,
	)
}

func (s *PingService) Ping(device string) error {
	id := proto.GenerateId()
	s.pings[id] = time.Now()
	msg := proto.Message{
		Id:          id,
		Action:      "ping",
		Destination: device,
	}
	return s.proto.Publish(msg)
}

func main() {
	app, err := core.NewApp("stark")
	app.Must(err)

	ctx := app.NewContext()
	srv := NewPingService(ctx.Proto)
	ctx.Must(srv.Enable())

	for _ = range time.Tick(1 * time.Second) {
		ctx.Must(srv.Ping(""))
	}
}
