// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/xconstruct/stark/proto"
)

type WebSocketConn struct {
	conn *websocket.Conn
}

func (c *WebSocketConn) Write(msg proto.Message) error {
	if err := msg.IsValid(); err != nil {
		return err
	}
	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer w.Close()
	return json.NewEncoder(w).Encode(msg)
}

func (c *WebSocketConn) Read() (proto.Message, error) {
	var msg proto.Message
	_, r, err := c.conn.NextReader()
	if err != nil {
		return msg, err
	}
	if err := json.NewDecoder(r).Decode(&msg); err != nil {
		return msg, err
	}
	return msg, msg.IsValid()
}

func (c *WebSocketConn) Close() error {
	return c.conn.Close()
}

func (s *Server) handleStreamStark(w http.ResponseWriter, r *http.Request) {
	// Check authentication.
	name := s.checkAuthentication(r)
	if name == "" {
		w.WriteHeader(401)
		s.Client.Log("err", "websocket auth error for "+r.RemoteAddr)
		return
	}

	ws, err := s.websocket.Upgrade(w, r, nil)
	if err != nil {
		s.Client.Log("err", "websocket upgrade error: "+err.Error())
		return
	}

	defer ws.Close()
	s.Client.Log("info", "new websocket conn from "+r.RemoteAddr+" as "+name)

	c := &WebSocketConn{ws}
	err = s.Broker.ListenOnConn(c)
	s.Client.Log("info", "websocket from "+r.RemoteAddr+" closed: "+err.Error())
}
