// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package tests

import (
	"github.com/sarifsystems/sarif/sarif"
	. "github.com/smartystreets/goconvey/convey"
)

func ServiceSchedulerTest(tr *TestRunner) {
	Convey("should receive simple task", func() {
		tr.When(sarif.CreateMessage("schedule/duration", map[string]interface{}{
			"duration": "300ms",
		}))

		reply := tr.Expect()
		So(reply, ShouldBeAction, "schedule/created")
	})

	Convey("should receive complex task", func() {
		tr.When(sarif.CreateMessage("schedule/duration", map[string]interface{}{
			"duration": "100ms",
			"reply": sarif.Message{
				Action:      "push/text",
				Destination: tr.Id,
				Text:        "reminder finished",
			},
		}))

		reply := tr.Expect()
		So(reply, ShouldBeAction, "schedule/created")
	})

	Convey("should emit both tasks", func() {
		reply := tr.Expect()
		So(reply, ShouldBeAction, "push/text")
		So(reply.Text, ShouldEqual, "reminder finished")

		reply = tr.Expect()
		So(reply, ShouldBeAction, "schedule/finished")
		So(reply.Text, ShouldStartWith, "Reminder from")
	})
}
