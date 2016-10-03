// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package lua

import (
	"strings"
	"sync"

	"github.com/sarifsystems/sarif/pkg/luareflect"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/yuin/gopher-lua"
)

type Machine struct {
	*sarif.Client
	Lua *lua.LState

	StateLock    sync.Mutex
	OutputBuffer string
}

func NewMachine(c *sarif.Client) *Machine {
	return &Machine{
		Lua:    lua.NewState(),
		Client: c,
	}
}

func (m *Machine) Enable() error {
	m.StateLock.Lock()
	defer m.StateLock.Unlock()

	m.Lua.OpenLibs()
	m.Lua.SetGlobal("print", m.Lua.NewFunction(m.luaPrint))

	mod := m.Lua.RegisterModule("sarif", map[string]lua.LGFunction{
		"subscribe":   m.luaSubscribe,
		"publish":     m.luaPublish,
		"request":     m.luaRequest,
		"natural":     m.luaNatural,
		"reply":       m.luaReply,
		"reply_error": m.luaReplyError,
	})
	m.Lua.SetField(mod, "device_id", lua.LString(m.DeviceId))
	if err := m.PreloadModuleString("fun", ModFun); err != nil {
		return err
	}
	if err := m.PreloadModuleString("store", ModStore); err != nil {
		return err
	}
	return nil
}

func (m *Machine) PreloadModuleString(name, source string) error {
	ls := m.Lua
	loader, err := ls.LoadString(source)
	if err != nil {
		return err
	}
	preload := ls.GetField(ls.GetField(ls.Get(lua.EnvironIndex), "package"), "preload")
	if _, ok := preload.(*lua.LTable); !ok {
		ls.RaiseError("package.preload must be a table")
	}
	ls.SetField(preload, name, loader)
	return nil
}

func (m *Machine) Disable() error {
	m.Lua.Close()
	return m.Disconnect()
}

func (m *Machine) moduleLoader(L *lua.LState) int {
	return 1
}

func (m *Machine) luaPrint(L *lua.LState) int {
	top := L.GetTop()
	for i := 1; i <= top; i++ {
		m.OutputBuffer += L.Get(i).String()
		if i != top {
			m.OutputBuffer += " "
		}
	}
	m.OutputBuffer += "\n"
	return 0
}

func (m *Machine) luaSubscribe(L *lua.LState) int {
	action := L.ToString(1)
	device := L.ToString(2)
	handler := L.ToFunction(3)

	m.Subscribe(action, device, func(msg sarif.Message) {
		m.luaHandle(msg, handler)
	})
	return 0
}

func (m *Machine) luaPublish(L *lua.LState) int {
	L.CheckTable(1)
	msg := tableToMessage(L, L.ToTable(1))

	if msg.Action != "" {
		m.Publish(msg)
	}
	return 0
}

func (m *Machine) luaReply(L *lua.LState) int {
	L.CheckTable(1)
	L.CheckTable(2)
	msg := tableToMessage(m.Lua, L.ToTable(1))
	reply := tableToMessage(m.Lua, L.ToTable(2))

	m.Reply(msg, reply)
	return 0
}

func (m *Machine) luaReplyError(L *lua.LState) int {
	L.CheckTable(1)
	typ := L.ToString(2)
	text := L.ToString(3)
	msg := tableToMessage(m.Lua, L.ToTable(1))

	m.Reply(msg, sarif.Message{
		Action: "err/" + typ,
		Text:   text,
	})
	return 0
}

func (m *Machine) luaRequest(L *lua.LState) int {
	L.CheckTable(1)
	msg := tableToMessage(m.Lua, L.ToTable(1))

	if msg.Action != "" {
		if reply, ok := <-m.Request(msg); ok {
			L.Push(messageToTable(L, reply))
			return 1
		}
	}
	return 0
}

func (m *Machine) luaNatural(L *lua.LState) int {
	L.CheckString(1)
	msg := sarif.Message{
		Action: "natural/handle",
		Text:   L.ToString(1),
	}

	if reply, ok := <-m.Request(msg); ok {
		L.Push(lua.LString(reply.Text))
		L.Push(messageToTable(L, reply))
		return 2
	}
	return 0
}

func (m *Machine) luaHandle(msg sarif.Message, handler *lua.LFunction) {
	m.StateLock.Lock()
	defer m.StateLock.Unlock()

	v := messageToTable(m.Lua, msg)
	err := m.Lua.CallByParam(lua.P{
		Fn:      handler,
		NRet:    1,
		Protect: true,
	}, v)
	if err != nil {
		m.Log("err", "handle err: "+err.Error())
	}

	for _, l := range strings.Split(m.FlushOut(), "\n") {
		if l != "" {
			m.Log("info", "print: "+l)
		}
	}
}

func (m *Machine) FlushOut() string {
	out := strings.TrimSpace(m.OutputBuffer)
	m.OutputBuffer = ""
	return out
}

func (m *Machine) Do(code string, arg interface{}) (string, error, interface{}) {
	m.StateLock.Lock()
	defer m.StateLock.Unlock()

	m.FlushOut()
	fn, err := m.Lua.LoadString(code)
	if err != nil {
		return "", err, nil
	}
	m.Lua.Push(fn)
	m.Lua.Push(luareflect.ToLua(m.Lua, arg))
	if err := m.Lua.PCall(1, 1, nil); err != nil {
		out := m.FlushOut()
		return out, err, nil
	}
	out := m.FlushOut()
	ret := m.Lua.Get(-1)
	m.Lua.Pop(1)

	var rv interface{}
	if ret != nil && ret != lua.LNil {
		rv = luareflect.DecodeToBasic(ret)
	}

	return out, nil, rv
}
