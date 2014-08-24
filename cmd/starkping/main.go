package main

import (
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
)

func main() {
	ctx, err := core.NewContext("stark")
	ctx.Must(err)

	pings := make(map[string]time.Time)

	cl := ctx.NewProtoClient("starkping")
	cl.RegisterHandler(func(msg proto.Message) {
		if msg.Action == "ping" {
			ctx.Must(cl.Publish(proto.Message{
				Action: "ack",
				CorrId: msg.Id,
			}))
		} else if msg.Action == "ack" {
			sent, ok := pings[msg.CorrId]
			if !ok {
				return
			}

			ctx.Log.Printf("%s from %s: time=%.1fms",
				msg.Action,
				msg.Source,
				time.Since(sent).Seconds()*1e3,
			)
		}
	})
	ctx.Must(cl.SubscribeSelf("ack"))

	for now := range time.Tick(1 * time.Second) {
		id := proto.GenerateId()
		pings[id] = now
		msg := proto.Message{
			Id:     id,
			Action: "ping",
		}
		ctx.Must(cl.Publish(msg))
	}

	select {}
}
