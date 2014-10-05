// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/xconstruct/stark/log"
	"github.com/xconstruct/stark/proto"
)

type Context struct {
	Database *DB
	Orm      *Orm
	Log      *log.Logger
	Proto    proto.Endpoint
	Config   *Config
}

func (ctx *Context) Must(err error) {
	if err != nil {
		ctx.Log.Fatalln(err)
	}
}

func NewTestContext() (*Context, proto.Endpoint) {
	var err error
	ctx := &Context{}
	ctx.Config = NewConfig()
	ctx.Log = log.Default

	if ctx.Orm, err = OpenDatabaseInMemory(); err != nil {
		log.Default.Fatalln(err)
	}
	ctx.Database = ctx.Orm.Database()

	a, b := proto.NewPipe()
	ctx.Proto = a

	return ctx, b
}
