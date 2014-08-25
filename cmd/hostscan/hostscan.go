// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

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
