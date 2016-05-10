// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// +build js

// A limited GopherJS demo version.
package main

import (
	"encoding/json"
	"log"

	"github.com/gopherjs/gopherjs/js"
	"github.com/xconstruct/stark/pkg/inject"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

type Server struct {
	Broker *proto.Broker
	*services.ModuleManager
}

func NewServer() *Server {
	s := &Server{
		Broker: proto.NewBroker(),
	}
	s.ModuleManager = services.NewModuleManager(s.instantiate)
	return s
}

type nullConfig struct{}

func (nullConfig) Exists() bool                  { return false }
func (nullConfig) Set(interface{}) error         { return nil }
func (nullConfig) Get(interface{}) (error, bool) { return nil, true }
func (nullConfig) Dir() string                   { return "" }

func (s *Server) instantiate(m *services.Module) (interface{}, error) {
	inj := inject.NewInjector()
	inj.Instance(s.Broker)
	inj.Factory(func() services.Config {
		return nullConfig{}
	})
	inj.Factory(func() proto.Conn {
		return s.Broker.NewLocalConn()
	})
	inj.Factory(func() *proto.Client {
		c := proto.NewClient("js/" + m.Name)
		c.Connect(s.Broker.NewLocalConn())
		return c
	})
	return inj.Create(m.NewInstance)
}

type Socket struct {
	conn   proto.Conn
	object *js.Object
}

func (s *Socket) Send(raw string) bool {
	var msg proto.Message
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		log.Println(err)
		return true
	}
	s.conn.Write(msg)
	return false
}

func (s *Socket) readLoop() {
	s.object.Set("readyState", 1)
	s.object.Call("onopen")

	for {
		msg, err := s.conn.Read()
		if err != nil {
			log.Println(err)
			continue
		}
		raw, err := json.Marshal(msg)
		if err != nil {
			log.Println(err)
			continue
		}
		m := js.Global.Get("Object").New()
		m.Set("data", string(raw))
		s.object.Call("onmessage", m)
	}
}

func (s *Server) NewSocketConn() *js.Object {
	conn := s.Broker.NewLocalConn()

	sock := &Socket{conn: conn}
	sock.object = js.MakeWrapper(sock)
	sock.object.Set("send", sock.object.Get("Send"))
	sock.object.Set("readyState", 0)
	go sock.readLoop()

	return sock.object
}
