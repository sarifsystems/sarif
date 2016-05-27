// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package meals

import (
	"testing"

	"github.com/sarifsystems/sarif/core"
	"github.com/sarifsystems/sarif/pkg/testutils"
	"github.com/sarifsystems/sarif/sarif"
)

var pizza = Product{
	Name:          "Generic Brand Pizza",
	ServingWeight: 415 * Gram,
	Stats: Stats{
		Weight: 415 * Gram,
		Energy: 3911 * Kilojoule,

		Fat:           0 * Gram,
		Carbohydrates: 113 * Gram,
		Sugar:         14 * Gram,
		Protein:       38 * Gram,
		Salt:          6.2 * Gram,
	},
}

var milk = Product{
	Name:          "Generic Moo Cow Milk",
	ServingVolume: 250 * Millilitre,
	Stats: Stats{
		Volume: 1 * Litre,
		Energy: 2670 * Kilojoule,

		Fat:           35 * Gram,
		Carbohydrates: 48 * Gram,
		Sugar:         48 * Gram,
		Protein:       33 * Gram,
		Salt:          1.3 * Gram,
	},
}

func TestCalculations(t *testing.T) {
	var total Stats
	total.Add(pizza.Servings(2))
	total.Add(milk.Servings(1))
	diff := total.Energy - 8489.5*Kilojoule
	if diff < -0.001 || diff > 0.001 {
		t.Error("unexpected total energy:", total.Energy)
	}
}

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

	st.Describe("Meals service", func() {

		st.It("should store a pizza", func() {
			st.When(sarif.CreateMessage("meal/product/new", pizza))
			st.ExpectAction("meal/product/created")
		})

		st.It("should store milk", func() {
			st.When(sarif.CreateMessage("meal/product/new", milk))
			st.ExpectAction("meal/product/created")
		})

		st.It("should record a serving", func() {
			st.When(sarif.CreateMessage("meal/record", map[string]interface{}{
				"size": 0.5,
				"name": "pizza",
			}))

			st.Expect(func(msg sarif.Message) {
				if !msg.IsAction("meal/serving/recorded") {
					t.Fatal("unexpected message received")
				}
				var sv Serving
				if err := msg.DecodePayload(&sv); err != nil {
					t.Fatal(err)
				}
				diff := sv.Stats().Energy - pizza.Energy*0.5
				if diff < -0.001 || diff > 0.001 {
					t.Error("wrong serving energy:", sv.Stats().Energy.StringKcal())
				}
			})
		})

		st.It("should fail when ambiguous", func() {
			st.When(sarif.CreateMessage("meal/record", map[string]interface{}{
				"size": 0.5,
				"name": "generic",
			}))

			st.ExpectAction("err/badrequest")
			st.ExpectAction("proto/log/err/badrequest")
		})

		st.It("should record text only", func() {
			st.When(sarif.Message{
				Action: "meal/record",
				Text:   "1 milk",
			})

			st.ExpectAction("meal/serving/recorded")
		})

		st.It("should provide daily stats", func() {
			st.When(sarif.CreateMessage("meal/stats", nil))

			st.Expect(func(msg sarif.Message) {
				if !msg.IsAction("meal/stats") {
					t.Fatal("unexpected message received")
				}
			})
		})
	})
}
