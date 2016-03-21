// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package scheduler

import (
	"strings"
	"testing"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/pkg/testutils"
	"github.com/xconstruct/stark/proto"
)

func TestService(t *testing.T) {
	t.Skip() // TODO: broken until I figure out integration tests with multiple services

	st := testutils.New(t)
	// setup context
	deps := &Dependencies{}
	st.UseConn(core.InjectTest(deps))

	// init service
	srv := NewService(deps)
	if err := srv.Enable(); err != nil {
		t.Fatal(err)
	}
	st.Wait()

	st.Describe("Scheduler", func() {

		st.It("should receive simple task", func() {
			st.When(proto.CreateMessage("schedule/duration", map[string]interface{}{
				"duration": "300ms",
			}))

			st.ExpectAction("schedule/created")
		})

		st.It("should receive complex task", func() {
			st.When(proto.CreateMessage("schedule/duration", map[string]interface{}{
				"duration": "100ms",
				"reply": proto.Message{
					Action: "push/text",
					Text:   "reminder finished",
				},
			}))

			st.ExpectAction("schedule/created")
		})

		st.It("should emit both tasks", func() {
			st.Expect(func(msg proto.Message) {
				if msg.Action != "push/text" {
					t.Error("did not receive scheduler reply")
				}
				if msg.Text != "reminder finished" {
					t.Error("did not receive correct payload:", msg.Text)
				}
			})

			st.Expect(func(msg proto.Message) {
				if msg.Action != "schedule/finished" {
					t.Error("did not receive scheduler reply")
				}
				if !strings.HasPrefix(msg.Text, "Reminder from") {
					t.Error("did not receive correct payload")
				}
			})
		})
	})
}
