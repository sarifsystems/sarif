// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package schema_test

import (
	"testing"

	"github.com/sarifsystems/sarif/pkg/schema"
)

type MyThing struct {
	schema.Thing
	Value int
}

type MySpecialThing struct {
	*schema.Thing `schema:"http://schema.org/Thing"`
	Name          string
}

func TestFill(t *testing.T) {
	my := &MyThing{Value: 3}
	schema.Fill(my)
	if my.SchemaType != "http://github.com/sarifsystems/sarif/pkg/schema_test/MyThing" {
		t.Error("wrong type:", my.SchemaType)
	}

	my2 := MySpecialThing{Name: "Shiny"}
	schema.Fill(&my2)
	if my2.SchemaType != "http://schema.org/Thing" {
		t.Error("wrong type:", my2.SchemaType)
	}
}
