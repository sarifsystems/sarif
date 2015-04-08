// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"log"

	"github.com/xconstruct/stark/pkg/inject"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

func InjectTest(container interface{}) proto.Conn {
	orm := OpenDatabaseInMemory()
	a, b := proto.NewPipe()
	broker := proto.NewBroker()
	broker.SetLogger(DefaultLog)
	go broker.ListenOnBridge(a)

	DefaultLog.SetLevel(LogLevelTrace)

	inj := inject.NewInjector()
	inj.Instance(orm.DB)
	inj.Instance(orm.Database())
	inj.Instance(proto.Conn(a))
	inj.Instance(broker)
	inj.Factory(func() services.Config {
		return NewConfig("").Section("test")
	})
	inj.Factory(func() proto.Logger {
		return DefaultLog
	})
	inj.Factory(func() *proto.Client {
		c := proto.NewClient("testclient-" + proto.GenerateId())
		c.Connect(broker.NewLocalConn())
		c.SetLogger(DefaultLog)
		return c
	})
	if err := inj.Inject(container); err != nil {
		log.Fatalln(err)
	}

	return b
}
