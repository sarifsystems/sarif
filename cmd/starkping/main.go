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

	ctx.Must(ctx.Proto.Subscribe("ack", func(msg proto.Message) {
		sent, ok := pings[msg.CorrId]
		if !ok {
			return
		}

		ctx.Log.Printf("%s from %s: time=%.1fms",
			msg.Action,
			msg.Source,
			time.Since(sent).Seconds()*1e3,
		)
	}))

	for now := range time.Tick(1 * time.Second) {
		id := proto.GenerateId()
		pings[id] = now
		msg := proto.Message{
			Id:     id,
			Action: "ping",
		}
		err := ctx.Proto.Publish(msg)
		ctx.Must(err)
	}

	select {}
}
