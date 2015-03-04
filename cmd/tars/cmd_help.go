// Copyright (C) 2015 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
)

const Usage = `Usage: tars [OPTION]... [MESSAGE]...

Natural commandline interface to the stark .
Publishes a natural message on the stark network, prints the response
and returns.

If invoked with no arguments, it starts an interactive session, where
each line from stdin published as a natural message.

Options:

`

func (app *App) Help() {
	topic := ""
	if flag.NArg() > 1 {
		topic = flag.Arg(1)
	}

	for _, c := range app.Commands {
		if c.Name == topic {
			log.Println(c.Help)
			return
		}
	}

	flag.Usage()
}
