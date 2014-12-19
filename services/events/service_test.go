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
				"subject": "user",
				"verb":    "drink",
				"object":  "coffee",
				"text":    "User drinks coffee.",
			}))

			st.ExpectAction("event/created")
		})

		st.It("should return last event", func() {
			st.When(proto.CreateMessage("event/last", map[string]interface{}{
				"verb": "drink",
			}))

			st.Expect(func(msg proto.Message) {
				if msg.Action != "event/found" {
					t.Error("did not find event")
				}
				got := Event{}
				msg.DecodePayload(&got)
				if got.Object != "coffee" {
					t.Error("did not find coffee")
				}
			})
		})

		st.It("should count events in a timeframe", func() {
			// Create test events
			st.When(proto.CreateMessage("event/new", map[string]interface{}{
				"subject":   "user",
				"verb":      "sleep",
				"status":    "started",
				"timestamp": time.Now().Add(-100 * time.Minute),
			}))
			st.ExpectAction("event/created")
			st.When(proto.CreateMessage("event/new", map[string]interface{}{
				"subject":   "user",
				"verb":      "sleep",
				"status":    "ended",
				"timestamp": time.Now().Add(-10 * time.Minute),
			}))
			st.ExpectAction("event/created")
			st.When(proto.CreateMessage("event/new", map[string]interface{}{
				"subject":   "user",
				"verb":      "sleep",
				"status":    "started",
				"timestamp": time.Now().AddDate(0, 0, -5),
			}))
			st.ExpectAction("event/created")

			// Count events
			st.When(proto.CreateMessage("event/count", map[string]interface{}{
				"verb":  "sleep",
				"after": time.Now().AddDate(0, 0, -3),
			}))
			st.Expect(func(msg proto.Message) {
				count := countPayload{}
				msg.DecodePayload(&count)
				if count.Count != 2 {
					t.Error("Wrong count: ", count.Count)
				}
			})

			// Summarize total duration
			st.When(proto.CreateMessage("event/sum/duration", map[string]interface{}{
				"verb":  "sleep",
				"after": time.Now().AddDate(0, 0, -3),
			}))
			st.Expect(func(msg proto.Message) {
				dur := sumDurationPayload{}
				msg.DecodePayload(&dur)
				if dur.Duration < 89*time.Minute || dur.Duration > 91*time.Minute {
					t.Error("Wrong duration: ", dur.Duration)
				}
			})
		})
	})
}
