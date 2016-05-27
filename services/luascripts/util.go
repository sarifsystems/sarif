// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package luascripts

import (
	"github.com/sarifsystems/sarif/pkg/luareflect"
	"github.com/sarifsystems/sarif/sarif"
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

func messageToTable(L *lua.LState, msg sarif.Message) lua.LValue {
	t := L.NewTable()
	tableSetString(t, "sarif", msg.Version)
	tableSetString(t, "id", msg.Id)
	tableSetString(t, "action", msg.Action)
	tableSetString(t, "src", msg.Source)
	tableSetString(t, "dest", msg.Destination)
	tableSetString(t, "corr", msg.CorrId)
	tableSetString(t, "text", msg.Text)

	p := make(map[string]interface{})
	msg.DecodePayload(&p)
	t.RawSetH(lua.LString("p"), luareflect.ToLua(L, p))
	return t
}

func tableToMessage(L *lua.LState, t *lua.LTable) sarif.Message {
	msg := sarif.Message{}
	msg.Version = tableGetString(t, "sarif")
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
