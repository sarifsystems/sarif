// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package luascripts

import (
	"strings"

	"github.com/xconstruct/stark/proto"
	"github.com/yuin/gopher-lua"
)

func (s *Service) moduleLoader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"subscribe":   s.luaSubscribe,
		"publish":     s.luaPublish,
		"request":     s.luaRequest,
		"reply":       s.luaReply,
		"reply_error": s.luaReplyError,
	})
	L.SetField(mod, "device_id", lua.LString(s.DeviceId))
	L.Push(mod)
	return 1
}

func (s *Service) luaSubscribe(L *lua.LState) int {
	action := L.ToString(1)
	device := L.ToString(2)
	handler := L.ToFunction(3)

	s.Subscribe(action, device, func(msg proto.Message) {
		s.luaHandle(msg, handler)
	})
	return 0
}

func (s *Service) luaPublish(L *lua.LState) int {
	s.Lua.CheckTable(1)
	msg := tableToMessage(s.Lua, s.Lua.ToTable(1))

	if msg.Action != "" {
		s.Publish(msg)
	}
	return 0
}

func (s *Service) luaReply(L *lua.LState) int {
	s.Lua.CheckTable(1)
	s.Lua.CheckTable(2)
	msg := tableToMessage(s.Lua, s.Lua.ToTable(1))
	reply := tableToMessage(s.Lua, s.Lua.ToTable(2))

	s.Reply(msg, reply)
	return 0
}

func (s *Service) luaReplyError(L *lua.LState) int {
	s.Lua.CheckTable(1)
	typ := L.ToString(2)
	text := L.ToString(3)
	msg := tableToMessage(s.Lua, s.Lua.ToTable(1))

	s.Reply(msg, proto.Message{
		Action: "err/" + typ,
		Text:   text,
	})
	return 0
}

func (s *Service) luaPrint(L *lua.LState) int {
	top := L.GetTop()
	for i := 1; i <= top; i++ {
		s.OutputBuffer += L.Get(i).String()
		if i != top {
			s.OutputBuffer += " "
		}
	}
	s.OutputBuffer += "\n"
	return 0
}

func (s *Service) luaRequest(L *lua.LState) int {
	s.Lua.CheckTable(1)
	msg := tableToMessage(s.Lua, s.Lua.ToTable(1))

	if msg.Action != "" {
		if reply, ok := <-s.Request(msg); ok {
			L.Push(messageToTable(L, reply))
			return 1
		}
	}
	return 0
}

func (s *Service) luaHandle(msg proto.Message, handler *lua.LFunction) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()
	v := messageToTable(s.Lua, msg)
	err := s.Lua.CallByParam(lua.P{
		Fn:      handler,
		NRet:    1,
		Protect: true,
	}, v)
	if err != nil {
		s.Log.Warnln("[luascripts] handle err:", err)
	}

	for _, l := range strings.Split(s.OutputBuffer, "\n") {
		if l != "" {
			s.Log.Infoln("[luascripts] print:", l)
		}
	}
}
