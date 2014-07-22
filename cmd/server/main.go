package main

import (
	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/services/hostscan"
)

func main() {
	ctx, err := core.NewContext()
	ctx.Must(err)

	h, err := hostscan.NewService(ctx)
	ctx.Must(err)
	ctx.Must(h.Enable())

	select {}
}
