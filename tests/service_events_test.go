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

type Event struct {
	Text string
}

func ServiceEventsTest(tr *TestRunner) {
	Convey("should store a new event", func() {
		tr.When(sarif.CreateMessage("event/new", map[string]interface{}{
			"action": "user/drink/coffee",
			"text":   "User drinks coffee.",
		}))

		reply := tr.Expect()
		So(reply, ShouldBeAction, "event/created")
	})

	Convey("should store return last event", func() {
		tr.When(sarif.CreateMessage("event/last", map[string]interface{}{
			"action": "user/drink/coffee",
		}))

		reply := tr.Expect()
		So(reply, ShouldBeAction, "event/found")

		payload := Event{}
		reply.DecodePayload(&payload)
		So(payload.Text, ShouldEqual, "User drinks coffee.")
	})

	Convey("should record other messages", func() {
		// Create test events
		tr.When(sarif.CreateMessage("event/record", map[string]interface{}{
			"action": "some/value/changed",
			"time":   time.Now().Add(-100 * time.Minute),
		}))
		So(tr.Expect(), ShouldBeAction, "event/recording")

		tr.When(sarif.Message{
			Action: "some/value/changed",
			Text:   "some value has changed",
		})
		tr.Wait()

		tr.When(sarif.CreateMessage("event/last", map[string]interface{}{
			"action": "some/value/changed",
		}))

		reply := tr.Expect()
		So(reply, ShouldBeAction, "event/found")

		payload := Event{}
		reply.DecodePayload(&payload)
		So(payload.Text, ShouldEqual, "some value has changed")
	})
}
