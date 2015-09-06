// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/xconstruct/stark/pkg/datasets/commands"
	"github.com/xconstruct/stark/pkg/natural"
)

func main() {
	set, err := natural.ReadDataSet(strings.NewReader(commands.Data))
	if err != nil {
		log.Fatal(err)
	}

	p := natural.NewVarPredictor()
	p.Train(10, set)

	pos, _ := strconv.Atoi(os.Args[2])
	guess, w := p.Predict(os.Args[1], "", pos)
	spew.Dump(guess, w)
}
