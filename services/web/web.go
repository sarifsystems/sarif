// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package web

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"crypto/rand"

	"code.google.com/p/go.net/websocket"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
)

const (
	REST_URL = "/api/v0/"
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
	Interface      string
	ApiKeys        map[string]string
	AllowedActions map[string][]string
}

type Server struct {
	cfg        Config
	ctx        *core.Context
	proto      *proto.Mux
	apiClients map[string]*proto.Client
}

func GenerateApiKey() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func New(ctx *core.Context) (*Server, error) {
	cfg := Config{
		Interface:      "0.0.0.0:5000",
		ApiKeys:        nil,
		AllowedActions: make(map[string][]string),
	}
	if err := ctx.Config.Get("web", &cfg); err != nil {
		return nil, err
	}
	if cfg.ApiKeys == nil {
		cfg.ApiKeys = make(map[string]string)
		cfg.ApiKeys["unprivileged"] = ""
		for i := 1; i < 6; i++ {
			key, err := GenerateApiKey()
			if err != nil {
				return nil, err
			}
			cfg.ApiKeys["exampleclient"+strconv.Itoa(i)] = key
		}
		cfg.AllowedActions["exampleclient1"] = []string{"ping", "location/update"}
		if err := ctx.Config.Set("web", cfg); err != nil {
			return nil, err
		}
	}

	s := &Server{
		cfg,
		ctx,
		proto.NewMux(),
		make(map[string]*proto.Client),
	}
	proto.Connect(ctx.Proto, s.proto)
	return s, nil
}

func NewInstance(ctx *core.Context) (core.ModuleInstance, error) {
	return New(ctx)
}

func (s *Server) Enable() error {
	http.Handle("/", http.FileServer(http.Dir("assets/web")))
	http.HandleFunc(REST_URL, s.handleRestPublish)
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
	s.ctx.Log.Infoln("[web] new websocket connection")

	// Check authentication.
	name := s.checkAuthentication(ws.Request())
	if name == "" {
		ws.WriteClose(401)
		return
	}

	mtp := s.proto.NewEndpoint()
	webtp := proto.NewByteEndpoint(ws)
	webtp.RegisterHandler(func(msg proto.Message) {
		s.ctx.Log.Debugln("[web] websocket received", msg)
		if err := mtp.Publish(msg); err != nil {
			s.ctx.Log.Errorln("[web] broker publish error:", err)
		}
	})
	mtp.RegisterHandler(func(msg proto.Message) {
		s.ctx.Log.Debugln("[web] mtp received:", msg)
		if err := webtp.Publish(msg); err != nil {
			s.ctx.Log.Errorln("[web] websocket publish error:", err)
		}
	})
	err := webtp.Listen()
	s.ctx.Log.Errorln("[web] websocket closed: ", err)
}

func parseAuthorizationHeader(h string) string {
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	switch parts[0] {
	case "Bearer":
		return parts[1]
	case "Basic":
		payload, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return ""
		}
		userpass := strings.SplitN(string(payload), ":", 2)
		return userpass[0]
	}
	return ""
}

func (s *Server) getApiClientByName(name string) *proto.Client {
	client, ok := s.apiClients[name]
	if !ok {
		client = proto.NewClient(name, s.proto.NewEndpoint())
		s.apiClients[name] = client
	}
	return client
}

func (s *Server) checkAuthentication(req *http.Request) string {
	fmt.Println(req)
	// Get authorization token.
	token := ""
	if auth := req.Header.Get("Authorization"); auth != "" {
		token = parseAuthorizationHeader(auth)
	}

	// Find client to API key.
	for name, stored := range s.cfg.ApiKeys {
		if token == stored {
			s.ctx.Log.Debugf("[web] authenticated for '%s'", name)
			return name
		}
	}
	s.ctx.Log.Warnln("[web] authentication failed")
	return ""
}

func (s *Server) clientIsAllowed(client string, msg proto.Message) bool {
	allowed, ok := s.cfg.AllowedActions[client]
	if !ok {
		return true
	}
	for _, action := range allowed {
		if msg.IsAction(action) {
			return true
		}
	}
	return false
}

func (s *Server) handleRestPublish(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	s.ctx.Log.Debugln("[web] new REST request:", req.URL.Path)

	// Parse form values.
	if err := req.ParseForm(); err != nil {
		s.ctx.Log.Warnln("[web] REST bad request:", err)
		w.WriteHeader(400)
		fmt.Fprintln(w, "Bad request:", err)
		return
	}

	// Check authentication.
	name := s.checkAuthentication(req)
	if name == "" {
		w.WriteHeader(401)
		fmt.Fprintln(w, "Not authorized")
		return
	}
	client := s.getApiClientByName(name)

	// Create message from form values.
	msg := proto.Message{
		Id:     proto.GenerateId(),
		Source: name,
	}
	if strings.HasPrefix(req.URL.Path, REST_URL) {
		msg.Action = strings.TrimPrefix(req.URL.Path, REST_URL)
	}
	pl := make(map[string]interface{})
	for k, v := range req.Form {
		if len(v) == 1 {
			pl[k] = v[0]
		} else {
			pl[k] = v
		}
	}
	_ = msg.EncodePayload(pl)

	if !s.clientIsAllowed(name, msg) {
		w.WriteHeader(401)
		fmt.Fprintf(w, "'%s' is not authorized to publish '%s'", name, msg.Action)
		s.ctx.Log.Warnf("[web] REST '%s' is not authorized to publish on '%s'", name, msg.Action)
		return
	}

	// Publish message.
	if err := client.Publish(msg); err != nil {
		s.ctx.Log.Warnln("[web] REST bad request:", err)
		w.WriteHeader(400)
		fmt.Fprintln(w, "Bad Request:", err)
		return
	}
	w.Write([]byte(msg.Id))
}
