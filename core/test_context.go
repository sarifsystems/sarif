// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"log"

	"github.com/sarifsystems/sarif/pkg/inject"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
)

func InjectTest(container interface{}) sarif.Conn {
	orm := OpenDatabaseInMemory()
	a, b := sarif.NewPipe()
	broker := sarif.NewBroker()
	broker.SetLogger(DefaultLog)
	go broker.ListenOnBridge(a)

	DefaultLog.SetLevel(LogLevelTrace)

	inj := inject.NewInjector()
	inj.Instance(orm.DB)
	inj.Instance(orm.Database())
	inj.Instance(sarif.Conn(a))
	inj.Instance(broker)
	inj.Factory(func() services.Config {
		return NewConfig("").Section("test")
	})
	inj.Factory(func() sarif.Logger {
		return DefaultLog
	})
	inj.Factory(func() *sarif.Client {
		c := sarif.NewClient("testclient-" + sarif.GenerateId())
		c.Connect(broker.NewLocalConn())
		c.SetLogger(DefaultLog)
		return c
	})
	if err := inj.Inject(container); err != nil {
		log.Fatalln(err)
	}

	return b
}
