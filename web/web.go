// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package web

import (
	"net/http"

	"code.google.com/p/go.net/websocket"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
)

var Module = core.Module{
	Name:        "web",
	Version:     "1.0",
	NewInstance: NewInstance,
}

func init() {
	core.RegisterModule(Module)
}

type Config struct {
	Interface string
}

type Server struct {
	cfg Config
	ctx *core.Context
}

func New(ctx *core.Context) (*Server, error) {
	cfg := Config{
		Interface: "0.0.0.0:5000",
	}
	err := ctx.Config.Get("web", &cfg)
	s := &Server{
		cfg,
		ctx,
	}
	return s, err
}

func NewInstance(ctx *core.Context) (core.ModuleInstance, error) {
	return New(ctx)
}

func (s *Server) Enable() error {
	http.Handle("/", http.FileServer(http.Dir("assets/web")))
	http.Handle("/stream/stark", websocket.Handler(s.handleStreamStark))

	go func() {
		s.ctx.Log.Infof("[web] listening on %s", s.cfg.Interface)
		err := http.ListenAndServe(s.cfg.Interface, nil)
		s.ctx.Log.Warnln(err)
	}()
	return nil
}

func (s *Server) Disable() error {
	return nil
}

func (s *Server) handleStreamStark(ws *websocket.Conn) {
	defer ws.Close()
	mtp := s.ctx.Proto.NewEndpoint()
	s.ctx.Log.Infoln("[web-socket] new connection")

	webtp := proto.NewByteEndpoint(ws)
	webtp.RegisterHandler(func(msg proto.Message) {
		if err := mtp.Publish(msg); err != nil {
			s.ctx.Log.Errorln("[web-mtp] ", err)
		}
	})
	mtp.RegisterHandler(func(msg proto.Message) {
		if err := webtp.Publish(msg); err != nil {
			s.ctx.Log.Errorln("[web-socket] ", err)
		}
	})
	err := webtp.Listen()
	s.ctx.Log.Errorln("[web-socket] closed: ", err)
}
