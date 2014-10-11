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
	deps := &Dependencies{}
	core.InjectTest(deps)

	db := &sqlDatabase{deps.DB.Driver(), deps.DB.DB}
	if err := db.Setup(); err != nil {
		t.Fatal(err)
	}

	err := db.StoreLocation(Location{
		Latitude:  52.3744779,
		Longitude: 9.7385532,
		Accuracy:  10,
		Source:    "Hannover",
	})
	if err != nil {
		t.Fatal(err)
	}

	g := Geofence{}
	g.SetBounds([]float64{52, 53, 9, 10})
	last, err := db.GetLastLocationInGeofence(g)
	if err != nil {
		t.Error(err)
	}
	if last.Source != "Hannover" {
		t.Errorf("Unexpected location: %s", last.Source)
	}

	g = Geofence{}
	g.SetBounds([]float64{52, 53, 10, 11})
	_, err = db.GetLastLocationInGeofence(g)
	if err != sql.ErrNoRows {
		t.Error(err)
	}
}

func TestService(t *testing.T) {
	deps := &Dependencies{}
	ep := core.InjectTest(deps)

	var lastReply *proto.Message
	ep.RegisterHandler(func(msg proto.Message) {
		lastReply = &msg
	})

	srv := NewService(deps)
	if err := srv.Enable(); err != nil {
		t.Fatal(err)
	}

	err := ep.Publish(proto.CreateMessage("location/update", map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"latitude":  52.3744779,
		"longitude": 9.7385532,
		"accuracy":  10,
		"source":    "Hannover",
	}))
	if err != nil {
		t.Fatal(err)
	}

	err = ep.Publish(proto.CreateMessage("location/last", map[string]interface{}{
		"bounds": []float64{52, 53, 9, 10},
	}))
	if err != nil {
		t.Fatal(err)
	}

	got := struct {
		Source string
	}{}
	lastReply.DecodePayload(&got)
	t.Log(lastReply, got)
	if got.Source != "Hannover" {
		t.Errorf("Unexpected location: %s", got.Source)
	}

	lastReply = nil

	err = ep.Publish(proto.CreateMessage("location/last", map[string]interface{}{
		"address": "Hannover, Germany",
	}))
	if err != nil {
		t.Fatal(err)
	}

	got = struct {
		Source string
	}{}
	lastReply.DecodePayload(&got)
	t.Log(lastReply, got)
	if got.Source != "Hannover" {
		t.Errorf("Unexpected location: %s", got.Source)
	}
}

func TestGeofence(t *testing.T) {
	deps := &Dependencies{}
	ep := core.InjectTest(deps)

	var lastReply proto.Message
	ep.RegisterHandler(func(msg proto.Message) {
		lastReply = msg
	})

	srv := NewService(deps)
	if err := srv.Enable(); err != nil {
		t.Fatal(err)
	}

	err := ep.Publish(proto.CreateMessage("location/fence/create", map[string]interface{}{
		"name":    "City",
		"lat_min": 5.1,
		"lat_max": 5.3,
		"lng_min": 6.1,
		"lng_max": 6.3,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !lastReply.IsAction("location/fence/created") {
		t.Fatal("expected a successful fence, not:", lastReply)
	}

	// outside of the fence
	err = ep.Publish(proto.CreateMessage("location/update", map[string]interface{}{
		"latitude":  5.2,
		"longitude": 6.0,
		"accuracy":  20,
	}))
	if err != nil {
		t.Fatal(err)
	}
	// inside of the fence
	err = ep.Publish(proto.CreateMessage("location/update", map[string]interface{}{
		"latitude":  5.2,
		"longitude": 6.2,
		"accuracy":  20,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !lastReply.IsAction("location/fence/enter") {
		t.Error("expected fence/enter, not:", lastReply)
	}

	// still inside
	err = ep.Publish(proto.CreateMessage("location/update", map[string]interface{}{
		"latitude":  5.2,
		"longitude": 6.2,
		"accuracy":  20,
	}))
	lastReply = proto.Message{}
	if err != nil {
		t.Fatal(err)
	}
	if lastReply.Action != "" {
		t.Error("expected no message, but got", lastReply)
	}

	// back outside
	lastReply = proto.Message{}
	err = ep.Publish(proto.CreateMessage("location/update", map[string]interface{}{
		"latitude":  5.4,
		"longitude": 6.0,
		"accuracy":  20,
	}))
	if err != nil {
		t.Fatal(err)
	}

	if !lastReply.IsAction("location/fence/leave") {
		t.Error("expected fence/leave, not:", lastReply)
	}
}
