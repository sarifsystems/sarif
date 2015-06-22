// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package events

import (
	"testing"
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/pkg/testutils"
	"github.com/xconstruct/stark/proto"
)

func TestService(t *testing.T) {
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
			st.When(proto.CreateMessage("event/new", map[string]interface{}{
				"action": "user/drink/coffee",
				"text":   "User drinks coffee.",
			}))

			st.ExpectAction("event/created")
		})

		st.It("should return last event", func() {
			st.When(proto.CreateMessage("event/last", map[string]interface{}{
				"action": "user/drink/coffee",
			}))

			st.Expect(func(msg proto.Message) {
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
			st.When(proto.CreateMessage("event/new", map[string]interface{}{
				"action": "user/sleep/start",
				"time":   time.Now().Add(-100 * time.Minute),
			}))
			st.ExpectAction("event/created")
			st.When(proto.CreateMessage("event/new", map[string]interface{}{
				"action": "user/sleep/end",
				"time":   time.Now().Add(-10 * time.Minute),
			}))
			st.ExpectAction("event/created")
			st.When(proto.CreateMessage("event/new", map[string]interface{}{
				"action": "user/sleep/start",
				"time":   time.Now().AddDate(0, 0, -5),
			}))
			st.ExpectAction("event/created")

			// Count events
			st.When(proto.CreateMessage("event/count", map[string]interface{}{
				"action": "user/sleep",
				"after":  time.Now().AddDate(0, 0, -3),
			}))
			st.Expect(func(msg proto.Message) {
				count := aggPayload{}
				msg.DecodePayload(&count)
				if count.Value != 2 {
					t.Error("Wrong count: ", count.Value)
				}
			})

			// Summarize total duration
			st.When(proto.CreateMessage("event/sum/duration", map[string]interface{}{
				"action": "user/sleep",
				"after":  time.Now().AddDate(0, 0, -3),
			}))
			st.Expect(func(msg proto.Message) {
				dur := sumDurationPayload{}
				msg.DecodePayload(&dur)
				d, ok := dur.Durations["user/sleep/start"]
				if !ok {
					t.Log(dur)
					t.Fatal("expected duration for 'start'")
				}
				if d < 89*60 || d > 91*60 {
					t.Error("Wrong duration: ", d)
				}
			})
		})

		st.It("should record other messages", func() {
			st.When(proto.CreateMessage("event/record", map[string]interface{}{
				"action": "some/value/changed",
			}))
			st.ExpectAction("event/recording")

			st.When(proto.Message{
				Action: "some/value/changed",
				Text:   "some value has changed",
			})

			st.When(proto.CreateMessage("event/last", map[string]interface{}{
				"action": "some/value/changed",
			}))

			st.Expect(func(msg proto.Message) {
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
