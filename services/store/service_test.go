// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package store

import (
	"testing"

	"github.com/sarifsystems/sarif/core"
	"github.com/sarifsystems/sarif/pkg/testutils"
	"github.com/sarifsystems/sarif/sarif"
)

type Person struct {
	Name      string  `json:"name"`
	Age       int     `json:"age"`
	Relevance float64 `json:"relevance"`
}

func TestService(t *testing.T) {
	t.Skip() // TODO: integration tests with database driver

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

	st.Describe("Store service", func() {

		st.It("should store a new document", func() {
			st.When(sarif.CreateMessage("store/put/person/john_smith", Person{
				Name:      "John Smith",
				Age:       43,
				Relevance: 1337.5443,
			}))

			st.ExpectAction("store/updated/person/john_smith") // reply
			st.ExpectAction("store/updated/person/john_smith") // publish
		})

		st.It("should return doc", func() {
			st.When(sarif.CreateMessage("store/get/person/john_smith", nil))

			st.Expect(func(msg sarif.Message) {
				st.ExpectAction("store/retrieved/person/john_smith")
				got := Person{}
				msg.DecodePayload(&got)
				if got.Name != "John Smith" {
					t.Error("did not find document")
				}
			})
		})

		st.It("should store another document", func() {
			st.When(sarif.CreateMessage("store/put/person/bob_benson", Person{
				Name:      "Bob Benson",
				Age:       27,
				Relevance: 50.3,
			}))

			st.ExpectAction("store/updated/person/bob_benson") // reply
			st.ExpectAction("store/updated/person/bob_benson") // publish
		})

		st.It("should scan all documents", func() {
			st.When(sarif.CreateMessage("store/scan/person", nil))

			st.Expect(func(msg sarif.Message) {
				st.ExpectAction("store/scanned/person")
				got := map[string]Person{}
				msg.DecodePayload(&got)
				if len(got) != 2 {
					t.Logf("%+v", got)
					t.Error("expected 2 results")
				}
				if p := got["john_smith"]; p.Age != 43 {
					t.Error("expected john smith")
				}
				if p := got["bob_benson"]; p.Age != 27 {
					t.Error("expected bob benson")
				}
			})
		})

		st.It("should support prefix scan", func() {
			st.When(sarif.CreateMessage("store/scan/person/john", nil))

			st.Expect(func(msg sarif.Message) {
				st.ExpectAction("store/scanned/person")
				got := map[string]Person{}
				msg.DecodePayload(&got)
				if len(got) != 1 {
					t.Logf("%+v", got)
					t.Error("expected 1 result")
				}
				if p := got["john_smith"]; p.Name != "John Smith" {
					t.Error("expected john smith")
				}
			})
		})

		st.It("should support reverse range", func() {
			st.When(sarif.CreateMessage("store/scan/person", map[string]interface{}{
				"end":     "zachary",
				"reverse": true,
				"limit":   1,
			}))

			st.Expect(func(msg sarif.Message) {
				st.ExpectAction("store/scanned/person")
				got := map[string]Person{}
				msg.DecodePayload(&got)
				if len(got) != 1 {
					t.Logf("%+v", got)
					t.Error("expected 1 result")
				}
				if p := got["john_smith"]; p.Name != "John Smith" {
					t.Error("expected john smith")
				}
			})
		})

		st.It("should support filtering", func() {
			st.When(sarif.CreateMessage("store/scan/person", map[string]interface{}{
				"filter": map[string]interface{}{
					"name ^":      "Bobs",
					"relevance <": 60,
				},
			}))

			st.Expect(func(msg sarif.Message) {
				st.ExpectAction("store/scanned/person")
				got := map[string]Person{}
				msg.DecodePayload(&got)
				if len(got) != 1 {
					t.Logf("%+v", got)
					t.Error("expected 1 result")
				}
				if p := got["bob_benson"]; p.Name != "Bob Benson" {
					t.Error("expected bob benson")
				}
			})
		})
	})
}
