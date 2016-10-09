// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package tests

import (
	"github.com/sarifsystems/sarif/sarif"
	. "github.com/smartystreets/goconvey/convey"
)

type Person struct {
	Name      string  `json:"name"`
	Age       int     `json:"age"`
	Relevance float64 `json:"relevance"`
}

func ServiceStoreTest(tr *TestRunner) {
	Convey("should store a new document", func() {
		tr.When(sarif.CreateMessage("store/put/person/john_smith", Person{
			Name:      "John Smith",
			Age:       43,
			Relevance: 1337.5443,
		}))

		reply := tr.Expect()
		So(reply, ShouldBeAction, "store/updated/person/john_smith")
	})

	Convey("should return the document", func() {
		tr.When(sarif.CreateMessage("store/get/person/john_smith", nil))

		reply := tr.Expect()
		So(reply, ShouldBeAction, "store/retrieved/person/john_smith")

		payload := Person{}
		reply.DecodePayload(&payload)
		So(payload.Name, ShouldEqual, "John Smith")
	})

	Convey("should return another document", func() {
		tr.When(sarif.CreateMessage("store/put/person/bob_benson", Person{
			Name:      "Bob Benson",
			Age:       27,
			Relevance: 50.3,
		}))

		reply := tr.Expect()
		So(reply, ShouldBeAction, "store/updated/person/bob_benson")
	})

	Convey("should scan all documents", func() {
		tr.When(sarif.CreateMessage("store/scan/person", nil))

		reply := tr.Expect()
		So(reply, ShouldBeAction, "store/scanned/person")
		got := make(map[string]interface{})
		reply.DecodePayload(&got)
		So(got, ShouldHaveLength, 2)

		So(got["keys"], ShouldResemble, []interface{}{"bob_benson", "john_smith"})
		So(got["values"], ShouldHaveLength, 2)
	})

	Convey("should support prefix scan", func() {
		tr.When(sarif.CreateMessage("store/scan/person/john", nil))

		reply := tr.Expect()
		So(reply, ShouldBeAction, "store/scanned/person")
		got := make(map[string]interface{})
		reply.DecodePayload(&got)
		So(got["keys"], ShouldResemble, []interface{}{"john_smith"})
	})

	Convey("should support reverse scan", func() {
		tr.When(sarif.CreateMessage("store/scan/person", map[string]interface{}{
			"end":     "zachary",
			"reverse": true,
			"limit":   1,
		}))

		reply := tr.Expect()
		So(reply, ShouldBeAction, "store/scanned/person")
		got := make(map[string]interface{})
		reply.DecodePayload(&got)
		So(got["keys"], ShouldResemble, []interface{}{"john_smith"})
	})

	Convey("should support filtering", func() {
		tr.When(sarif.CreateMessage("store/scan/person", map[string]interface{}{
			"filter": map[string]interface{}{
				"name ^":      "Bob",
				"relevance <": 60,
			},
			"only": "keys",
		}))

		reply := tr.Expect()
		So(reply, ShouldBeAction, "store/scanned/person")
		got := make([]interface{}, 0)
		reply.DecodePayload(&got)
		So(got, ShouldResemble, []interface{}{"bob_benson"})
	})
}
