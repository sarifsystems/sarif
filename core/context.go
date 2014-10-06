// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/xconstruct/stark/proto"
)

type Context struct {
	Database *DB
	Orm      *Orm
	Log      *Logger
	Proto    proto.Conn
	Config   *Config
}

func (ctx *Context) Must(err error) {
	if err != nil {
		ctx.Log.Fatalln(err)
	}
}

func NewTestContext() (*Context, proto.Conn) {
	var err error
	ctx := &Context{}
	ctx.Config = NewConfig()
	ctx.Log = DefaultLog

	if ctx.Orm, err = OpenDatabaseInMemory(); err != nil {
		ctx.Log.Fatalln(err)
	}
	ctx.Database = ctx.Orm.Database()

	a, b := proto.NewPipe()
	ctx.Proto = a

	return ctx, b
}
