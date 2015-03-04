// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package inject

import (
	"log"
	"os"
	"testing"
)

type testType struct {
	Value string
}

type dependencies struct {
	Log  *log.Logger
	Type *testType
}

type instance struct {
	Type *testType
}

func newInstance(deps *dependencies) *instance {
	return &instance{
		deps.Type,
	}
}

func TestInject(t *testing.T) {
	in := NewInjector()
	in.Instance(log.New(os.Stderr, "test ", 0))
	in.Factory(func() *testType {
		return &testType{"moo"}
	})

	deps := &dependencies{}
	if err := in.Inject(deps); err != nil {
		t.Fatal(err)
	}
	if deps.Log == nil {
		t.Error("Expected logger instance")
	}
	if deps.Type == nil {
		t.Error("Expected testType instance")
	}

	deps.Log.Println("weee")
	if deps.Type.Value != "moo" {
		t.Error("Wrong value:", deps.Type.Value)
	}

	intf, err := in.Create(newInstance)
	if err != nil {
		t.Fatal(err)
	}
	instance, ok := intf.(*instance)
	if !ok {
		t.Fatal("Could not convert %t to instance", intf)
	}
	if instance.Type == nil {
		t.Error("Expected testType instance")
	}
	if instance.Type.Value != "moo" {
		t.Error("Wrong value:", instance.Type.Value)
	}
}
