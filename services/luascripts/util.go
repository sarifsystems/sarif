// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package luascripts

import (
	"github.com/xconstruct/stark/pkg/luareflect"
	"github.com/xconstruct/stark/proto"
	"github.com/yuin/gopher-lua"
)

func tableSetString(t *lua.LTable, key, val string) {
	if val != "" {
		t.RawSetH(lua.LString(key), lua.LString(val))
	}
}

func tableGetString(t *lua.LTable, key string) string {
	s := lua.LVAsString(t.RawGetH(lua.LString(key)))
	if s == "nil" {
		return ""
	}
	return s
}

func messageToTable(L *lua.LState, msg proto.Message) lua.LValue {
	t := L.NewTable()
	tableSetString(t, "stark", msg.Version)
	tableSetString(t, "id", msg.Id)
	tableSetString(t, "action", msg.Action)
	tableSetString(t, "src", msg.Source)
	tableSetString(t, "dest", msg.Destination)
	tableSetString(t, "corr", msg.CorrId)
	tableSetString(t, "text", msg.Text)

	p := make(map[string]interface{})
	msg.DecodePayload(p)
	t.RawSetH(lua.LString("p"), luareflect.ToLua(L, p))
	return t
}

func tableToMessage(L *lua.LState, t *lua.LTable) proto.Message {
	msg := proto.Message{}
	msg.Version = tableGetString(t, "stark")
	msg.Id = tableGetString(t, "id")
	msg.Action = tableGetString(t, "action")
	msg.Source = tableGetString(t, "src")
	msg.Destination = tableGetString(t, "dest")
	msg.CorrId = tableGetString(t, "corr")
	msg.Text = tableGetString(t, "text")

	p := luareflect.DecodeToBasic(t.RawGetH(lua.LString("p")))
	msg.EncodePayload(p)
	return msg
}
