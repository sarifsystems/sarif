// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package luareflect implements simple mappings from Go to Gopher-Lua types.
package luareflect

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func ToLua(L *lua.LState, data interface{}) lua.LValue {
	if data == nil {
		return lua.LNil
	}
	val := reflect.ValueOf(data)
	kind := getKind(val)

	switch kind {
	case reflect.Bool:
		return lua.LBool(val.Bool())
	case reflect.String:
		return lua.LString(val.String())
	case reflect.Int:
		return lua.LNumber(val.Int())
	case reflect.Uint:
		return lua.LNumber(val.Uint())
	case reflect.Float32:
		return lua.LNumber(val.Float())
	case reflect.Ptr:
		return ToLua(L, reflect.Indirect(val).Interface())
	case reflect.Slice:
		t := L.NewTable()
		for i := 0; i < val.Len(); i++ {
			t.Append(ToLua(L, val.Index(i).Interface()))
		}
		return t
	case reflect.Struct:
		typ := val.Type()
		t := L.NewTable()
		for i := 0; i < typ.NumField(); i++ {
			lk := lua.LString(typ.Field(i).Name)
			lv := ToLua(L, val.Field(i).Interface())
			t.RawSet(lk, lv)
		}
		return t
	case reflect.Map:
		t := L.NewTable()
		for _, k := range val.MapKeys() {
			lk := ToLua(L, k.Interface())
			lv := ToLua(L, val.MapIndex(k).Interface())
			t.RawSet(lk, lv)
		}
		return t
	}
	return lua.LNil
}

func DecodeToBasic(data lua.LValue) interface{} {
	switch data.Type() {
	case lua.LTNil:
		return nil
	case lua.LTBool:
		return lua.LVAsBool(data)
	case lua.LTNumber:
		return float64(lua.LVAsNumber(data))
	case lua.LTString:
		return lua.LVAsString(data)
	case lua.LTTable:
		m := make(map[string]interface{})
		data.(*lua.LTable).ForEach(func(key, val lua.LValue) {
			if k := lua.LVAsString(key); k != "" {
				m[k] = DecodeToBasic(val)
			}
		})
		return m
	}
	return nil
}

func getKind(val reflect.Value) reflect.Kind {
	kind := val.Kind()
	switch {
	case kind >= reflect.Int && kind <= reflect.Int64:
		return reflect.Int
	case kind >= reflect.Uint && kind <= reflect.Uint64:
		return reflect.Uint
	case kind >= reflect.Float32 && kind <= reflect.Float64:
		return reflect.Float32
	default:
		return kind
	}
}
