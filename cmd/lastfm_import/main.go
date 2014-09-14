// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/log"
	"github.com/xconstruct/stark/services/lastfm"
)

var verbose = flag.Bool("v", false, "verbose debug output")

func main() {
	flag.Parse()
	app := core.NewApp("stark")
	app.Must(app.Init())
	defer app.Close()

	if *verbose {
		app.Log.SetLevel(log.LevelDebug)
	}

	ctx := app.NewContext()
	srv, err := lastfm.NewService(ctx)
	app.Must(err)
	app.Must(srv.Enable())
	app.Must(srv.ImportAll())
}
