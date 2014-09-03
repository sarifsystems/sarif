// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package events

import (
	"testing"

	"github.com/xconstruct/stark/core"
)

func TestStoreRetrieve(t *testing.T) {
	// setup test database
	ctx, _ := core.NewTestContext()
	db := &sqlDatabase{ctx.Database.Driver(), ctx.Database.DB}
	if err := db.Setup(); err != nil {
		t.Fatal(err)
	}

	// store two events
	err := db.StoreEvent(Event{
		Subject: "user",
		Verb:    "drink",
		Object:  "coffee",
		Status:  "singular",
		Text:    "User drinks coffee.",
		Meta: map[string]interface{}{
			"flavor": "Caramel Macchiato",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.StoreEvent(Event{
		Subject: "user",
		Verb:    "is_at_geofence",
		Object:  "home",
		Status:  "started",
		Text:    "User enters home.",
	})
	if err != nil {
		t.Fatal(err)
	}

	// retrieve an event (#2)
	event, err := db.GetLastEvent(Event{
		Verb:   "is_at_geofence",
		Object: "home",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(event)
	if event.Timestamp.IsZero() {
		t.Error("expected valid timestamp, not", event.Timestamp)
	}
	if event.Subject != "user" {
		t.Error("expected user as subject, not", event.Subject)
	}
	if event.Verb != "is_at_geofence" {
		t.Error("expected is_at_geofence as verb, not", event.Verb)
	}

	// retrieve an event (#1)
	event, err = db.GetLastEvent(Event{
		Subject: "user",
		Status:  "singular",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(event)
	if event.Timestamp.IsZero() {
		t.Error("expected valid timestamp, not", event.Timestamp)
	}
	if event.Object != "coffee" {
		t.Error("expected cofee as object, not", event.Object)
	}
	if event.Meta["flavor"] != "Caramel Macchiato" {
		t.Error("wrong coffee flavor")
	}
}
