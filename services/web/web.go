// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service web provides a web dashboard and communication between sarif and HTTP.
package web

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
)

const (
	REST_URL = "/api/v0/"
)

var Module = &services.Module{
	Name:        "web",
	Version:     "1.0",
	NewInstance: New,
}

type Config struct {
	Interface      string
	ApiKeys        map[string]string
	AllowedActions map[string][]string
}

type Dependencies struct {
	Config services.Config
	Broker *sarif.Broker
	Client *sarif.Client
}

type Server struct {
	Config     services.Config
	cfg        Config
	Broker     *sarif.Broker
	apiClients map[string]*sarif.Client
	Client     *sarif.Client
	websocket  websocket.Upgrader
}

func GenerateApiKey() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func New(deps *Dependencies) *Server {
	cfg := Config{
		Interface:      "0.0.0.0:5000",
		ApiKeys:        nil,
		AllowedActions: make(map[string][]string),
	}
	if deps.Config.Exists() {
		deps.Config.Get(&cfg)
	} else {
		cfg.ApiKeys = make(map[string]string)
		for i := 1; i < 6; i++ {
			key, err := GenerateApiKey()
			if err != nil {
				deps.Client.Log("err", err.Error())
				continue
			}
			cfg.ApiKeys["exampleclient"+strconv.Itoa(i)] = key
		}
		cfg.AllowedActions["exampleclient1"] = []string{"ping", "location/update"}
		deps.Config.Set(cfg)
	}

	s := &Server{
		Config:     deps.Config,
		cfg:        cfg,
		Broker:     deps.Broker,
		apiClients: make(map[string]*sarif.Client),
		Client:     deps.Client,
		websocket: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 2014,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
	return s
}

func (s *Server) Enable() error {
	dir := s.Config.Dir() + "/web"
	http.Handle("/", http.FileServer(http.Dir(dir)))
	http.HandleFunc(REST_URL, s.handleRestPublish)
	http.HandleFunc("/socket", s.handleSocket)

	s.Client.Subscribe("json", "", s.handleJson)

	go func() {
		s.Client.Log("info", "listening on "+s.cfg.Interface)
		err := http.ListenAndServe(s.cfg.Interface, nil)
		s.Client.Log("err", "http listen error: "+err.Error())
	}()
	return nil
}

func (s *Server) Disable() error {
	return nil
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

func (s *Server) getApiClientByName(name string) *sarif.Client {
	client, ok := s.apiClients[name]
	if !ok {
		client = sarif.NewClient("web/" + name)
		client.Connect(s.Broker.NewLocalConn())
		s.apiClients[name] = client
	}
	return client
}

func (s *Server) checkAuthentication(req *http.Request) string {
	// Get authorization token.
	token := ""
	if auth := req.Header.Get("Authorization"); auth != "" {
		token = parseAuthorizationHeader(auth)
	}
	if token == "" && req.FormValue("authtoken") != "" {
		token = req.FormValue("authtoken")
	}

	// Find client to API key.
	for name, stored := range s.cfg.ApiKeys {
		if token == stored {
			return name
		}
	}
	return ""
}

func (s *Server) clientIsAllowed(client string, msg sarif.Message) bool {
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
	s.Client.Log("debug", "new REST request: "+req.URL.Path)

	// Parse form values.
	if err := req.ParseForm(); err != nil {
		s.Client.Log("warn", "REST bad request: "+err.Error())
		w.WriteHeader(400)
		fmt.Fprintln(w, "Bad request:", err)
		return
	}

	// Check authentication.
	name := s.checkAuthentication(req)
	if name == "" {
		w.WriteHeader(401)
		fmt.Fprintln(w, "Not authorized")
		s.Client.Log("info", "authentication failed for "+req.RemoteAddr)
		return
	}
	s.Client.Log("info", "authenticated "+req.RemoteAddr+" for "+name)
	client := s.getApiClientByName(name)

	// Create message from form values.
	msg := sarif.Message{
		Id:     sarif.GenerateId(),
		Source: name,
	}
	if strings.HasPrefix(req.URL.Path, REST_URL) {
		msg.Action = strings.TrimPrefix(req.URL.Path, REST_URL)
	}
	pl := make(map[string]interface{})
	for k, v := range req.Form {
		if k == "authtoken" {
			continue
		}
		if k == "_device" {
			msg.Destination = v[0]
		} else if k == "text" {
			msg.Text = strings.Join(v, "\n")
		} else if len(v) == 1 {
			pl[k] = parseValue(v[0])
		} else if k == "_device" {
			pl[k] = v
		}
	}
	_ = msg.EncodePayload(pl)

	if !s.clientIsAllowed(name, msg) {
		w.WriteHeader(401)
		fmt.Fprintf(w, "'%s' is not authorized to publish '%s'", name, msg.Action)
		s.Client.Log("warn", "REST '"+name+"' is not authorized to publish on "+msg.Action)
		return
	}

	// Publish message.
	if err := client.Publish(msg); err != nil {
		s.Client.Log("warn", "REST bad request: "+err.Error())
		w.WriteHeader(400)
		fmt.Fprintln(w, "Bad Request:", err)
		return
	}
	w.Write([]byte(msg.Id))
}

func parseValue(s string) interface{} {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}
	return s
}
