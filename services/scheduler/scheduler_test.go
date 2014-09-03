// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package scheduler

import (
	"strings"
	"testing"
	"time"

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

	// send simple default task
	err = ep.Publish(proto.Message{
		Action: "schedule/duration",
		Payload: map[string]interface{}{
			"duration": "300ms",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// wait for confirmation
	t.Log("confirmation:", lastReply)
	if lastReply.Action != "schedule/created" {
		t.Error("did not receive confirmation for creation")
	}

	lastReply = nil
	// send task with payload
	err = ep.Publish(proto.Message{
		Action: "schedule/duration",
		Payload: map[string]interface{}{
			"duration": "100ms",
			"reply": proto.Message{
				Action: "push/text",
				Payload: map[string]interface{}{
					"text": "reminder finished",
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// wait for confirmation
	t.Log("confirmation:", lastReply)
	if lastReply.Action != "schedule/created" {
		t.Error("did not receive confirmation for creation")
	}
	lastReply = nil

	// wait for task with payload to fire
	time.Sleep(200 * time.Millisecond)
	t.Log("reply:", lastReply)
	if lastReply.Action != "push/text" {
		t.Error("did not receive scheduler reply")
	}
	if v := lastReply.PayloadGetString("text"); v != "reminder finished" {
		t.Error("did not receive correct payload:", v)
	}

	// wait for simple task to fire
	time.Sleep(200 * time.Millisecond)
	t.Log("reply:", lastReply)
	if lastReply.Action != "schedule/finished" {
		t.Error("did not receive scheduler reply")
	}
	if !strings.HasPrefix(lastReply.PayloadGetString("text"), "Reminder from") {
		t.Error("did not receive correct payload")
	}
}
