// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package trigger

import (
	"testing"
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/pkg/testutils"
	"github.com/xconstruct/stark/proto"
)

const eventGeofenceTemplate = `
{
	"action": "event/new",
	"p": {
		"subject": "user",
		"verb": "{{.status}}",
		"object": "{{.fence.name}}",
		"object_type": "geofence",
		"meta": {{.}},
		"status": "{{.status}}"
	}
}
`

func TestService(t *testing.T) {
	st := testutils.New(t)
	deps := &Dependencies{}
	st.UseConn(core.InjectTest(deps))
	st.WaitTimeout = 5 * time.Second

	// init service
	srv := NewService(deps)
	if err := srv.Enable(); err != nil {
		t.Fatal(err)
	}
	st.Wait()

	st.Describe("Trigger service", func() {

		st.It("should store a new rule", func() {
			st.When(proto.CreateMessage("trigger/new", map[string]interface{}{
				"name":   "emit_event_on_geofence",
				"action": "location/fence",
				"reply":  eventGeofenceTemplate,
			}))
			st.ExpectAction("trigger/created")
			st.Wait()
		})

		st.It("should apply stored rule", func() {
			st.When(proto.CreateMessage("location/fence/enter", map[string]interface{}{
				"status": "enter",
				"fence": map[string]interface{}{
					"name": "Home",
				},
			}))

			st.ExpectAction("event/new")
		})
	})
}
