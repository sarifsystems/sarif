// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package luascripts

import (
	"os"
	"strings"
	"sync"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/pkg/luareflect"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
	"github.com/yuin/gopher-lua"
)

var Module = &services.Module{
	Name:        "luascripts",
	Version:     "1.0",
	NewInstance: NewService,
}

type Config struct {
	ScriptDir string `json:"script_dir"`
}

type Dependencies struct {
	Config *core.Config
	Log    proto.Logger
	Client *proto.Client
}

type Service struct {
	cfg Config
	Log proto.Logger
	*proto.Client

	stateLock    sync.Mutex
	Lua          *lua.LState
	OutputBuffer string
}

func NewService(deps *Dependencies) *Service {
	s := &Service{
		Log:    deps.Log,
		Client: deps.Client,
		Lua:    nil,
	}
	deps.Config.Get("luascripts", &s.cfg)
	return s
}

func (s *Service) Enable() error {
	s.stateLock.Lock()
	s.Lua = lua.NewState()
	s.Lua.OpenLibs()
	s.Lua.SetGlobal("print", s.Lua.NewFunction(s.luaPrint))
	s.Lua.SetGlobal("subscribe", s.Lua.NewFunction(s.luaSubscribe))
	s.Lua.SetGlobal("publish", s.Lua.NewFunction(s.luaPublish))
	s.Lua.SetGlobal("request", s.Lua.NewFunction(s.luaRequest))
	s.stateLock.Unlock()

	s.Subscribe("lua/do", "", s.handleLuaDo)
	s.Subscribe("cmd/lua", "", s.handleLuaDo)

	if s.cfg.ScriptDir == "" {
		return nil
	}
	dir, err := os.Open(s.cfg.ScriptDir)
	if err != nil {
		return err
	}
	files, err := dir.Readdirnames(0)
	if err != nil {
		return err
	}
	for _, f := range files {
		if !strings.HasSuffix(f, ".lua") {
			continue
		}
		s.Log.Infoln("[luascripts] loading ", f)
		if err := s.Lua.DoFile(s.cfg.ScriptDir + "/" + f); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Disable() error {
	s.Lua.Close()
	return nil
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

func tableSetString(t *lua.LTable, key, val string) {
	if val != "" {
		t.RawSetH(lua.LString(key), lua.LString(val))
	}
}

func tableGetString(t *lua.LTable, key string) string {
	return lua.LVAsString(t.RawGetH(lua.LString(key)))
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

func (s *Service) handleLuaDo(msg proto.Message) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()
	s.OutputBuffer = ""
	if err := s.Lua.DoString(msg.Text); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	text := strings.TrimSpace(s.OutputBuffer)
	s.OutputBuffer = ""
	s.Reply(msg, proto.Message{
		Action: "lua/done",
		Text:   text,
	})
}
