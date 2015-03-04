// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"log"

	"github.com/xconstruct/stark/pkg/inject"
	"github.com/xconstruct/stark/proto"
)

func InjectTest(container interface{}) proto.Conn {
	orm := OpenDatabaseInMemory()
	a, b := proto.NewPipe()

	inj := inject.NewInjector()
	inj.Instance(orm.DB)
	inj.Instance(orm.Database())
	inj.Instance(proto.Conn(a))
	inj.Instance(NewConfig(""))
	inj.Factory(func() proto.Logger {
		return DefaultLog
	})
	inj.Factory(func() *proto.Client {
		c := proto.NewClient("testclient-" + proto.GenerateId())
		c.Connect(a)
		c.SetLogger(DefaultLog)
		return c
	})
	if err := inj.Inject(container); err != nil {
		log.Fatalln(err)
	}

	return b
}
