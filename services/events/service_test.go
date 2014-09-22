// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package events

import (
	"testing"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
)

func TestService(t *testing.T) {
	// setup context
	ctx, ep := core.NewTestContext()
	var lastReply *proto.Message
	ep.RegisterHandler(func(msg proto.Message) {
		lastReply = &msg
	})

	// init service
	srv, err := NewService(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err = srv.Enable(); err != nil {
		t.Fatal(err)
	}

	// store new event
	err = ep.Publish(proto.CreateMessage("event/new", map[string]interface{}{
		"subject": "user",
		"verb":    "drink",
		"object":  "coffee",
		"text":    "User drinks coffee.",
	}))
	if err != nil {
		t.Fatal(err)
	}

	// check confirmation
	t.Log("confirmation:", lastReply)
	if lastReply.Action != "event/created" {
		t.Error("did not receive confirmation for creation")
	}

	lastReply = nil
	// retrieve event
	err = ep.Publish(proto.CreateMessage("event/last", map[string]interface{}{
		"verb": "drink",
	}))
	if err != nil {
		t.Fatal(err)
	}

	// check reply
	t.Log("reply:", lastReply)
	if lastReply.Action != "event/found" {
		t.Error("did not find event")
	}
	got := Event{}
	lastReply.DecodePayload(&got)
	if got.Object != "coffee" {
		t.Error("did not find coffee")
	}
	lastReply = nil
}
