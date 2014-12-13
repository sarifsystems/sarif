package schema_test

import (
	"testing"

	"github.com/xconstruct/stark/pkg/schema"
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
	if my.SchemaType != "http://github.com/xconstruct/stark/pkg/schema_test/MyThing" {
		t.Error("wrong type:", my.SchemaType)
	}

	my2 := MySpecialThing{Name: "Shiny"}
	schema.Fill(&my2)
	if my2.SchemaType != "http://schema.org/Thing" {
		t.Error("wrong type:", my2.SchemaType)
	}
}
