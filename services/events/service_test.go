// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package events

import (
	"testing"
	"time"

	"github.com/sarifsystems/sarif/core"
	"github.com/sarifsystems/sarif/pkg/testutils"
	"github.com/sarifsystems/sarif/sarif"
)

func TestService(t *testing.T) {
	t.Skip() // TODO: broken until I figure out integration tests with multiple services

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

	st.Describe("Events service", func() {

		st.It("should store a new event", func() {
			st.When(sarif.CreateMessage("event/new", map[string]interface{}{
				"action": "user/drink/coffee",
				"text":   "User drinks coffee.",
			}))

			st.ExpectAction("event/created")
		})

		st.It("should return last event", func() {
			st.When(sarif.CreateMessage("event/last", map[string]interface{}{
				"action": "user/drink/coffee",
			}))

			st.Expect(func(msg sarif.Message) {
				st.ExpectAction("event/found")
				got := Event{}
				msg.DecodePayload(&got)
				if got.Text != "User drinks coffee." {
					t.Error("did not find coffee")
				}
			})
		})

		st.It("should count events in a timeframe", func() {
			// Create test events
			st.When(sarif.CreateMessage("event/new", map[string]interface{}{
				"action": "user/sleep/start",
				"time":   time.Now().Add(-100 * time.Minute),
			}))
			st.ExpectAction("event/created")
			st.When(sarif.CreateMessage("event/new", map[string]interface{}{
				"action": "user/sleep/end",
				"time":   time.Now().Add(-10 * time.Minute),
			}))
			st.ExpectAction("event/created")
			st.When(sarif.CreateMessage("event/new", map[string]interface{}{
				"action": "user/sleep/start",
				"time":   time.Now().AddDate(0, 0, -5),
			}))
			st.ExpectAction("event/created")

			// Count events
			st.When(sarif.CreateMessage("event/count", map[string]interface{}{
				"action": "user/sleep",
				"after":  time.Now().AddDate(0, 0, -3),
			}))
			st.Expect(func(msg sarif.Message) {
				count := aggPayload{}
				msg.DecodePayload(&count)
				if count.Value != 2 {
					t.Error("Wrong count: ", count.Value)
				}
			})
		})

		st.It("should record other messages", func() {
			st.When(sarif.CreateMessage("event/record", map[string]interface{}{
				"action": "some/value/changed",
			}))
			st.ExpectAction("event/recording")

			st.When(sarif.Message{
				Action: "some/value/changed",
				Text:   "some value has changed",
			})

			st.When(sarif.CreateMessage("event/last", map[string]interface{}{
				"action": "some/value/changed",
			}))

			st.Expect(func(msg sarif.Message) {
				st.ExpectAction("event/found")
				got := Event{}
				msg.DecodePayload(&got)
				if got.Text != "some value has changed" {
					t.Error("did not find text:", got.Text)
				}
			})
		})
	})
}
