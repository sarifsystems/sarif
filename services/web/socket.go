// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/transports/sfproto"
)

type WebSocketConn struct {
	conn  *websocket.Conn
	mutex sync.Mutex
}

func (c *WebSocketConn) Write(msg sarif.Message) error {
	if err := msg.IsValid(); err != nil {
		return err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	defer w.Close()
	return json.NewEncoder(w).Encode(msg)
}

func (c *WebSocketConn) Read() (sarif.Message, error) {
	var msg sarif.Message
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

func (s *Server) handleSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := s.websocket.Upgrade(w, r, nil)
	if err != nil {
		s.Client.Log("err", "websocket upgrade error: "+err.Error())
		return
	}

	defer ws.Close()
	s.Client.Log("info", "new websocket conn from "+r.RemoteAddr)

	c := &WebSocketConn{conn: ws}
	err = s.Broker.AuthenticateAndListenOnConn(sfproto.AuthChallenge, c)
	s.Client.Log("info", "websocket from "+r.RemoteAddr+" closed: "+err.Error())
}
