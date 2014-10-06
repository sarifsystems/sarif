// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/services/lastfm"
)

func main() {
	app := core.NewApp("stark")
	app.Must(app.Init())
	defer app.Close()

	ctx := app.NewContext()
	srv, err := lastfm.NewService(ctx)
	app.Must(err)
	app.Must(srv.Enable())
	app.Must(srv.ImportAll())
}
