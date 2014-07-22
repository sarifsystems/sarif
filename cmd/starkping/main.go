package main

import (
	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
)

func main() {
	ctx, err := core.NewContext()
	ctx.Must(err)

	c, err := ctx.Client()
	ctx.Must(err)

	ctx.Must(c.Publish(proto.Message{
		Action: "ping",
	}))

	select {}
}
