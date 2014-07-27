package main

import (
	"fmt"
	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/services/hostscan"
)

func main() {
	ctx, err := core.NewContext("stark")
	ctx.Must(err)

	db := ctx.Database

	ctx.Must(hostscan.SetupSchema(db.Driver(), db.DB))
	h := hostscan.New(db.DB)
	hosts, err := h.Update()
	ctx.Must(err)
	fmt.Println(hosts)
}
