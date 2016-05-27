// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/sarifsystems/sarif/core/server"
	"github.com/sarifsystems/sarif/services/lastfm"
)

func main() {
	app := server.New("sarif", "server")
	app.Init()
	defer app.Close()

	deps := &lastfm.Dependencies{}
	app.Must(app.Inject("lastfm", deps))
	srv := lastfm.NewService(deps)
	app.Must(srv.Enable())
	app.Must(srv.ImportAll())
}
