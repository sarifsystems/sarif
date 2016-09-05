// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build integration

package tests

import (
	"testing"

	"github.com/sarifsystems/sarif/core"
	"github.com/sarifsystems/sarif/core/server"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services/events"
	"github.com/sarifsystems/sarif/services/location"
	"github.com/sarifsystems/sarif/services/lua"
	"github.com/sarifsystems/sarif/services/scheduler"
	"github.com/sarifsystems/sarif/services/store"
	"github.com/smartystreets/goconvey/convey"

	_ "github.com/sarifsystems/sarif/services/store/bolt"
)

var runTests = []Test{
	{"Store Service", ServiceStoreTest},
	{"Scheduler Service", ServiceSchedulerTest},
	{"Events Service", ServiceEventsTest},
	{"Location Service", ServiceLocationTest},
	{"Lua Service", ServiceLuaTest},
}

type Test struct {
	Name string
	Func func(*TestRunner)
}

func TestSpec(t *testing.T) {
	srv := server.New("sarif", "temp")
	srv.Log.SetLevel(core.LogLevelWarn)

	srv.RegisterModule(events.Module)
	srv.RegisterModule(location.Module)
	srv.RegisterModule(lua.Module)
	srv.RegisterModule(scheduler.Module)
	srv.RegisterModule(store.Module)

	srv.ServerConfig.EnabledModules = []string{
		"events",
		"location",
		"lua",
		"scheduler",
		"store",
	}
	srv.Init()
	// srv.Log.SetLevel(core.LogLevelTrace)
	// srv.Broker.TraceMessages(true)

	c := sarif.NewClient("testservice")
	c.Connect(srv.Broker.NewLocalConn())
	c.SetLogger(srv.Log)

	for _, tst := range runTests {
		tr := NewTestRunner(t)
		tr.UseConn(srv.Broker.NewLocalConn())
		tr.Wait()

		convey.Convey("Test module '"+tst.Name+"'", t, func() {
			tst.Func(tr)
		})
	}

	srv.Close()
}
