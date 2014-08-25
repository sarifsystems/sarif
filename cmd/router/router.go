// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/services/router"
)

func main() {
	ctx, err := core.NewContext("stark")
	ctx.Must(err)

	var cfg router.Config
	ctx.Must(ctx.Config.Get("router", &cfg))

	r := router.New(cfg)

	ctx.Must(r.Login())

	diag, err := r.Diagnostic()
	fmt.Println(diag)
	ctx.Must(err)

	ctx.Must(r.Logout())
}
