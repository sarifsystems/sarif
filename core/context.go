// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"log"

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
	ctx := &Context{}
	ctx.Config = NewConfig()
	ctx.Log = DefaultLog

	ctx.Orm = OpenDatabaseInMemory()
	ctx.Database = ctx.Orm.Database()

	a, b := proto.NewPipe()
	ctx.Proto = a

	return ctx, b
}

func InjectTest(container interface{}) proto.Conn {
	orm := OpenDatabaseInMemory()
	a, b := proto.NewPipe()

	inj := NewInjector()
	inj.Instance(orm.DB)
	inj.Instance(orm.Database())
	inj.Instance(proto.Conn(a))
	inj.Instance(NewConfig())
	inj.Factory(func() proto.Logger {
		return DefaultLog
	})
	inj.Factory(func() *proto.Client {
		c := proto.NewClient("test", a)
		c.SetLogger(DefaultLog)
		return c
	})
	if err := inj.Inject(container); err != nil {
		log.Fatalln(err)
	}

	return b
}
