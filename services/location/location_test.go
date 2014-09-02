// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package location

import (
	"database/sql"
	"testing"
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
)

func TestStoreRetrieve(t *testing.T) {
	ctx, _ := core.NewTestContext()

	db := &sqlDatabase{ctx.Database.Driver(), ctx.Database.DB}
	if err := db.Setup(); err != nil {
		t.Fatal(err)
	}

	err := db.Store(Location{
		Latitude:  52.3744779,
		Longitude: 9.7385532,
		Accuracy:  10,
		Source:    "Hannover",
	})
	if err != nil {
		t.Fatal(err)
	}

	last, err := db.GetLastLocationInBounds(52, 53, 9, 10)
	if err != nil {
		t.Error(err)
	}
	if last.Source != "Hannover" {
		t.Errorf("Unexpected location: %s", last.Source)
	}

	_, err = db.GetLastLocationInBounds(52, 53, 10, 11)
	if err != sql.ErrNoRows {
		t.Error(err)
	}
}

func TestService(t *testing.T) {
	ctx, ep := core.NewTestContext()

	var lastReply *proto.Message
	ep.RegisterHandler(func(msg proto.Message) {
		lastReply = &msg
	})

	srv, err := NewService(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if err = srv.Enable(); err != nil {
		t.Fatal(err)
	}

	err = ep.Publish(proto.Message{
		Action: "location/update",
		Payload: map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"latitude":  52.3744779,
			"longitude": 9.7385532,
			"accuracy":  10,
			"source":    "Hannover",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = ep.Publish(proto.Message{
		Action: "location/last",
		Payload: map[string]interface{}{
			"bounds": []float64{52, 53, 9, 10},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(lastReply)
	if lastReply.PayloadGetString("source") != "Hannover" {
		t.Errorf("Unexpected location: %s", lastReply.PayloadGetString("source"))
	}
}
