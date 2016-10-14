// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package tests

import (
	"github.com/sarifsystems/sarif/sarif"
	. "github.com/smartystreets/goconvey/convey"
)

func ServiceNaturalTest(tr *TestRunner) {
	tr.Subscribe("reply/test")
	tr.Subscribe("question/answer/ultimate")

	Convey("should parse a simple command", func() {
		tr.When(sarif.Message{
			Action: "natural/parse",
			Text:   "/proto/ping",
		})

		reply := tr.Expect()
		So(reply, ShouldBeAction, "natural/parsed")
		got := struct {
			Intents []map[string]interface{} `json:"intents"`
		}{}
		reply.DecodePayload(&got)
		So(got.Intents, ShouldHaveLength, 1)
		So(got.Intents[0]["intent"], ShouldEqual, "proto/ping")
		So(got.Intents[0]["type"], ShouldEqual, "simple")
	})

	Convey("should learn a new association rule", func() {
		tr.When(sarif.Message{
			Action: "natural/handle",
			Text:   "associate reply [text] with /reply/test",
		})

		reply := tr.Expect()
		So(reply, ShouldBeAction, "natural/learned/meaning")
	})

	Convey("should react to learned rule", func() {
		tr.When(sarif.Message{
			Action: "natural/handle",
			Text:   "reply this is awesome",
		})

		reply := tr.Expect()
		So(reply, ShouldBeAction, "reply/test")
		So(reply.Text, ShouldEqual, "this is awesome")
	})

	Convey("should provide simple phrases", func() {
		tr.When(sarif.CreateMessage("natural/phrases/technobabble", nil))

		reply := tr.Expect()
		So(reply, ShouldBeAction, "natural/phrase")
	})

	Convey("should handle question and answers", func() {
		Convey("by first receiving the question", func() {
			msg := sarif.CreateMessage("question/ask", map[string]interface{}{
				"action": map[string]interface{}{
					"@type": "TextEntryAction",
					"reply": "question/answer/ultimate",
				},
			})
			msg.Text = "What is the answer to life, the universe and everything?"
			msg.Destination = "myserver/natural/" + tr.Id
			tr.When(msg)

			reply := tr.Expect()
			So(reply, ShouldBeAction, "question/ask")
		})

		Convey("and then answering it", func() {
			tr.When(sarif.Message{
				Action: "natural/handle",
				Text:   "42",
			})

			reply := tr.Expect()
			So(reply, ShouldBeAction, "question/answer/ultimate")
			So(reply.Text, ShouldEqual, "42")
		})
	})
}
