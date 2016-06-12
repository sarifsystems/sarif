// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package tests

import (
	"time"

	"github.com/sarifsystems/sarif/sarif"
	. "github.com/smartystreets/goconvey/convey"
)

func ServiceLocationTest(tr *TestRunner) {
	tr.Subscribe("location/fence/enter")
	tr.Subscribe("location/fence/leave")
	tr.Subscribe("location/fence/created")

	Convey("should store a location update", func() {
		tr.When(sarif.CreateMessage("location/update", map[string]interface{}{
			"timestamp": time.Now(),
			"latitude":  52.3744779,
			"longitude": 9.7385532,
			"accuracy":  10,
			"source":    "Hannover",
		}))
		tr.Wait()

		tr.When(sarif.CreateMessage("location/last", map[string]interface{}{
			"bounds": []float64{52, 53, 9, 10},
		}))

		reply := tr.Expect()
		So(reply, ShouldBeAction, "location/found")
		got := struct {
			Source string
		}{}
		reply.DecodePayload(&got)
		So(got.Source, ShouldEqual, "Hannover")
	})

	Convey("should answer a geocoded address", func() {
		tr.When(sarif.CreateMessage("location/last", map[string]interface{}{
			"address": "Hannover, Germany",
		}))

		reply := tr.Expect()
		got := struct {
			Source string
		}{}
		reply.DecodePayload(&got)
		So(got.Source, ShouldEqual, "Hannover")
	})

	Convey("should store a geofence", func() {
		tr.When(sarif.CreateMessage("location/fence/create", map[string]interface{}{
			"name":    "City",
			"lat_min": 5.1,
			"lat_max": 5.3,
			"lng_min": 6.1,
			"lng_max": 6.3,
		}))

		So(tr.Expect(), ShouldBeAction, "location/fence/created")
	})

	Convey("should emit a geofence enter event", func() {
		// outside of the fence
		tr.When(sarif.CreateMessage("location/update", map[string]interface{}{
			"latitude":  5.2,
			"longitude": 6.0,
			"accuracy":  20,
		}))
		tr.Wait()

		// inside of the fence
		tr.When(sarif.CreateMessage("location/update", map[string]interface{}{
			"latitude":  5.2,
			"longitude": 6.2,
			"accuracy":  20,
		}))

		So(tr.Expect(), ShouldBeAction, "location/fence/enter")
	})

	Convey("should emit a geofence leave event", func() {
		// still inside
		tr.When(sarif.CreateMessage("location/update", map[string]interface{}{
			"latitude":  5.2,
			"longitude": 6.2,
			"accuracy":  20,
		}))
		tr.Wait()

		// back outside
		tr.When(sarif.CreateMessage("location/update", map[string]interface{}{
			"latitude":  5.4,
			"longitude": 6.0,
			"accuracy":  20,
		}))

		So(tr.Expect(), ShouldBeAction, "location/fence/leave")
	})
}
