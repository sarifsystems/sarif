// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// A limited GopherJS demo version.
// +build js
package main

import (
	"log"

	"github.com/gopherjs/gopherjs/js"
	"github.com/xconstruct/stark/services/natural"
	"github.com/xconstruct/stark/services/nlparser"
)

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	srv := NewServer()

	srv.RegisterModule(natural.Module)
	srv.RegisterModule(nlparser.Module)

	must(srv.EnableModule("natural"))
	must(srv.EnableModule("nlparser"))

	js.Global.Set("StarkServer", js.MakeWrapper(srv))
}
