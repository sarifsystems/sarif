// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package tests

import (
	"github.com/sarifsystems/sarif/sarif"
	. "github.com/smartystreets/goconvey/convey"
)

func ServiceLuaTest(tr *TestRunner) {
	Convey("should execute a simple script", func() {
		tr.When(sarif.Message{
			Action: "lua/do",
			Text:   "print(3 + 5)",
		})

		reply := tr.Expect()
		So(reply, ShouldBeAction, "lua/done")
		So(reply.Text, ShouldEqual, "8")
	})

	Convey("should react to messages", func() {
		tr.When(sarif.Message{
			Action: "lua/do",
			Text: `
			sarif.subscribe("my/repeat", "", function(msg)
				sarif.reply(msg, {
					action = "my/repeated",
					text = msg.text .. msg.text,
				})
			end)
			`,
		})

		reply := tr.Expect()
		So(reply, ShouldBeAction, "lua/done")

		tr.Wait()
		tr.Wait()
		tr.Wait()

		tr.When(sarif.Message{
			Action: "my/repeat",
			Text:   "mooo",
		})

		reply = tr.Expect()
		So(reply, ShouldBeAction, "my/repeated")
		So(reply.Text, ShouldEqual, "mooomooo")
	})

	Convey("should request messages", func() {
		tr.Subscribe("my/request")
		tr.Subscribe("got")

		tr.When(sarif.Message{
			Action: "lua/do",
			Text: `
			sarif.subscribe("", "self", function() end)
			local rep = sarif.request{
				action = "my/request",
				text = "hello from inside",
			}
			sarif.publish{
				action = "got",
				text = rep.action .. ": " .. rep.text,
			}
			`,
		})

		reply := tr.Expect()
		So(reply, ShouldBeAction, "my/request")

		tr.Wait()
		tr.Wait()
		tr.Wait()

		tr.When(sarif.Message{
			Action:      "my/response",
			Destination: reply.Source,
			Text:        "hello from outside",
			CorrId:      reply.Id,
		})

		reply = tr.Expect()
		So(reply, ShouldBeAction, "lua/done")

		reply = tr.Expect()
		So(reply, ShouldBeAction, "got")
		So(reply.Text, ShouldEqual, "my/response: hello from outside")
	})
}
