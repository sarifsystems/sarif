package testutils

import (
	"testing"

	"github.com/xconstruct/stark/proto"
)

func TestUtils(t *testing.T) {
	st := New(t)

	// test service to test the tester
	conn := st.CreateConn()
	go func() {
		for {
			conn.Read()
			conn.Write(proto.Message{
				Action: "hi",
			})
			conn.Write(proto.Message{
				Action: "still/there",
			})
		}
	}()

	// Test cases
	st.Describe("My service", func() {
		st.It("Should reply", func() {
			st.When(proto.CreateMessage("hello", nil))

			st.Expect(func(msg proto.Message) {
				if !msg.IsAction("hi") {
					st.Fatal("expected hi, not ", msg.Action)
				}
			})

			st.Expect(func(msg proto.Message) {
				if !msg.IsAction("still/there") {
					st.Fatal("expected still/there, not ", msg.Action)
				}
			})
		})

		st.It("Should reply again", func() {
			st.When(proto.CreateMessage("hello", nil))

			st.ExpectAction("hi")
			st.ExpectAction("still/there")
		})
	})
}
