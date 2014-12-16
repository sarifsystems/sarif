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
	"github.com/xconstruct/stark/pkg/testutils"
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
	st := testutils.New(t)
	deps := &Dependencies{}
	st.UseConn(core.InjectTest(deps))
	st.WaitTimeout = 5 * time.Second

	// init service
	srv := NewService(deps)
	if err := srv.Enable(); err != nil {
		t.Fatal(err)
	}
	st.Wait()

	st.Describe("Location service", func() {

		st.It("should store a location update", func() {
			st.When(proto.CreateMessage("location/update", map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"latitude":  52.3744779,
				"longitude": 9.7385532,
				"accuracy":  10,
				"source":    "Hannover",
			}))

			st.When(proto.CreateMessage("location/last", map[string]interface{}{
				"bounds": []float64{52, 53, 9, 10},
			}))

			st.Expect(func(msg proto.Message) {
				got := struct {
					Source string
				}{}
				msg.DecodePayload(&got)
				if got.Source != "Hannover" {
					t.Errorf("Unexpected location: %s", got.Source)
				}
			})
		})

		st.It("should answer a geocoded address", func() {
			st.When(proto.CreateMessage("location/last", map[string]interface{}{
				"address": "Hannover, Germany",
			}))

			st.Expect(func(msg proto.Message) {
				got := struct {
					Source string
				}{}
				msg.DecodePayload(&got)
				if got.Source != "Hannover" {
					t.Errorf("Unexpected location: %s", got.Source)
				}
			})
		})
	})

	st.Describe("Geofence service", func() {

		st.It("should store a geofence", func() {
			st.When(proto.CreateMessage("location/fence/create", map[string]interface{}{
				"name":    "City",
				"lat_min": 5.1,
				"lat_max": 5.3,
				"lng_min": 6.1,
				"lng_max": 6.3,
			}))

			st.Expect(func(msg proto.Message) {
				if !msg.IsAction("location/fence/created") {
					t.Fatal("expected a successful fence, not:", msg)
				}
			})
		})

		st.It("should emit a geofence enter event", func() {
			// outside of the fence
			st.When(proto.CreateMessage("location/update", map[string]interface{}{
				"latitude":  5.2,
				"longitude": 6.0,
				"accuracy":  20,
			}))

			// inside of the fence
			st.When(proto.CreateMessage("location/update", map[string]interface{}{
				"latitude":  5.2,
				"longitude": 6.2,
				"accuracy":  20,
			}))

			st.Expect(func(msg proto.Message) {
				if !msg.IsAction("location/fence/enter") {
					t.Error("expected fence/enter, not:", msg)
				}
			})
		})

		st.It("should emit a geofence leave event", func() {
			// still inside
			st.When(proto.CreateMessage("location/update", map[string]interface{}{
				"latitude":  5.2,
				"longitude": 6.2,
				"accuracy":  20,
			}))

			// back outside
			st.When(proto.CreateMessage("location/update", map[string]interface{}{
				"latitude":  5.4,
				"longitude": 6.0,
				"accuracy":  20,
			}))

			st.Expect(func(msg proto.Message) {
				if !msg.IsAction("location/fence/leave") {
					t.Error("expected fence/enter, not:", msg)
				}
			})
		})
	})
}
