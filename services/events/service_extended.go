// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package events

import "github.com/xconstruct/stark/proto"

type locationFenceChanged struct {
	Status string `json:"status"`
	Fence  struct {
		Name string `json:"name"`
	} `json:"fence"`
}

func (s *Service) handleLocationFence(msg proto.Message) {
	var pl locationFenceChanged
	if err := msg.DecodePayload(&pl); err != nil {
		s.ctx.Log.Warnln("[events] received bad payload:", err)
		return
	}

	meta := make(map[string]interface{})
	msg.DecodePayload(&meta)
	e := Event{
		Subject: "user",
		Verb:    pl.Status + "_geofence",
		Object:  pl.Fence.Name,
		Meta:    meta,
		Text:    "User " + pl.Status + "s " + pl.Fence.Name + ".",
	}
	reply := proto.CreateMessage("event/new", e)
	s.proto.Publish(reply)
}
