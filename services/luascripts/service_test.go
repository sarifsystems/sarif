// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package luascripts

import (
	"testing"

	"github.com/sarifsystems/sarif/core"
	"github.com/sarifsystems/sarif/pkg/testutils"
	"github.com/sarifsystems/sarif/sarif"
)

func TestService(t *testing.T) {
	t.Skip("TODO: Broken because broker race conditions")

	// setup context
	st := testutils.New(t)
	deps := &Dependencies{}
	st.UseConn(core.InjectTest(deps))

	// init service
	srv := NewService(deps)
	if err := srv.Enable(); err != nil {
		t.Fatal(err)
	}
	st.Wait()

	st.Describe("Luascripts service", func() {

		st.It("should execute a simple script", func() {
			st.When(sarif.Message{
				Action: "lua/do",
				Text:   "print(3 + 5)",
			})
			st.Expect(func(msg sarif.Message) {
				st.ExpectAction("lua/done")
				st.ExpectText("8")
			})
		})

		st.It("should react to messages", func() {
			st.When(sarif.Message{
				Action: "lua/do",
				Text: `
				sarif.subscribe("my/repeat", "", function(msg)
					sarif.publish({
						action = "my/repeated",
						text = msg.text .. msg.text,
					})
				end)
				`,
			})
			st.ExpectAction("lua/done")

			st.When(sarif.Message{
				Action: "my/repeat",
				Text:   "mooo",
			})
			st.Expect(func(msg sarif.Message) {
				st.ExpectAction("my/repeated")
				st.ExpectText("mooomooo")
			})
		})

		st.It("should request messages", func() {
			st.When(sarif.Message{
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

			st.Expect(func(msg sarif.Message) {
				st.ExpectAction("my/request")
				st.When(sarif.Message{
					Action:      "my/response",
					Destination: msg.Source,
					Text:        "hello from outside",
					CorrId:      msg.Id,
				})
			})

			st.Expect(func(msg sarif.Message) {
				st.ExpectAction("got")
				st.ExpectText("my/response: hello from outside")
			})
			st.ExpectAction("lua/done")
		})
	})
}
