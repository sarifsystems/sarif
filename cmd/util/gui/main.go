// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"net/http"

	"github.com/miketheprogrammer/go-thrust/lib/commands"
	"github.com/miketheprogrammer/go-thrust/lib/connection"
	"github.com/miketheprogrammer/go-thrust/thrust"
	"github.com/sarifsystems/sarif/core"
	"github.com/sarifsystems/sarif/sarif"
	"golang.org/x/net/websocket"
)

func main() {
	app := New()
	app.Init()
	defer app.Close()

	go app.RunBackend()
	app.RunFrontend()
}

type WebClient struct {
	*core.App
}

func New() *WebClient {
	app := core.NewApp("sarif", "client")
	return &WebClient{
		app,
	}
}

func (c *WebClient) RunBackend() {
	http.Handle("/", http.FileServer(http.Dir("assets/web")))
	http.Handle("/stream/sarif", websocket.Handler(c.handleStreamSarif))
	err := http.ListenAndServe("localhost:54693", nil)
	c.Log.Fatal("backend:", err)
}

func (c *WebClient) RunFrontend() {
	thrust.InitLogger()
	thrust.Start()

	thrustWindow := thrust.NewWindow(thrust.WindowOptions{
		RootUrl: "http://localhost:54693/#/chat",
	})
	thrustWindow.Show()
	thrustWindow.Focus()

	quit := make(chan bool, 0)
	thrust.NewEventHandler("closed", func(cr commands.EventResult) {
		connection.CleanExit()
		close(quit)
	})
	<-quit
}

func (c *WebClient) handleStreamSarif(ws *websocket.Conn) {
	defer ws.Close()
	c.Log.Infoln("[web] new websocket connection")

	webtp := sarif.NewByteConn(ws)
	gateway := c.Dial()

	err := sarif.Transmit(webtp, gateway)
	c.Log.Errorln("[web] websocket closed:", err)
}
