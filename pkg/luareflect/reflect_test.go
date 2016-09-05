// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package luareflect

import (
	"reflect"
	"testing"

	"github.com/yuin/gopher-lua"
)

func TestBasicToLua(t *testing.T) {
	c := &Tester{t, nil}

	c.Set("my teststr")
	c.Run(`
		assert(v == "my teststr", "failed to encode string")
	`)

	c.Set(35)
	c.Run(`
		assert(v == 35, "failed to encode int")
	`)

	c.Set(1337.235)
	c.Run(`
		assert(v == 1337.235, "failed to encode float")
	`)

	c.Set(true)
	c.Run(`
		assert(v, "failed to encode bool")
	`)
}

func TestMapToLua(t *testing.T) {
	c := &Tester{t, nil}

	c.Set(map[string]interface{}{
		"a_string": "this is a string",
		"a_number": 123.456,
		"nested": map[int]interface{}{
			37: "thirty-seven",
			5:  "five",
		},
	})
	c.Run(`
		assert(type(v)      == "table")
		assert(v.a_string   == "this is a string")
		assert(v.a_number   == 123.456)
		assert(v.nested[37] == "thirty-seven")
		assert(v.nested[5]  == "five")
	`)
}

func TestSliceToLua(t *testing.T) {
	c := &Tester{t, nil}
	c.Set([]interface{}{3, "str", 54.7, true})
	c.Run(`
		assert(type(v)      == "table")
		assert(#v == 4)
		assert(v[1] == 3)
		assert(v[2] == "str")
		assert(v[3] == 54.7)
		assert(v[4] == true)
	`)
}

type Thing struct {
	Text   string
	Number float64
}

type MultiThing struct {
	Name string
	Thing
	Another *Thing
}

func TestStructToLua(t *testing.T) {
	c := &Tester{t, nil}

	c.Set(&MultiThing{
		Name:    "something",
		Thing:   Thing{"magic number", 256.513},
		Another: &Thing{"very special", 3},
	})
	c.Run(`
		assert(type(v) == "table")
		assert(v.Name == "something")
		assert(v.Thing.Text == "magic number")
		assert(v.Thing.Number == 256.513)
		assert(v.Another.Text == "very special")
		assert(v.Another.Number == 3)
	`)
}

type Tester struct {
	*testing.T
	V interface{}
}

func (t *Tester) Set(v interface{}) {
	t.V = v
}

func (t *Tester) Run(test string) {
	L := lua.NewState()
	defer L.Close()
	L.SetGlobal("v", ToLua(L, t.V))
	if err := L.DoString(test); err != nil {
		t.Fatal(err)
	}
	t.V = nil
}

func TestTableDecodeToBasic(t *testing.T) {
	L := lua.NewState()
	defer L.Close()
	v := ` v =
	{
		one = "two",
		[3] = 4,
		nested = {
			something = "value",
		},
		an_array = {
			"first",
			false,
		},
	}
	`
	if err := L.DoString(v); err != nil {
		t.Fatal(err)
	}
	got := DecodeToBasic(L.GetGlobal("v"))

	expected := map[string]interface{}{
		"one": "two",
		"3":   float64(4),
		"nested": map[string]interface{}{
			"something": "value",
		},
		"an_array": []interface{}{
			"first",
			false,
		},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Log(got)
		t.Fatal("error decoding table")
	}
}
