// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service xmpp provides access to the sarif network over XMPP.
package xmpp

import (
	"strings"
	"time"

	"github.com/agl/xmpp-client/xmpp"
	"github.com/sarifsystems/sarif/pkg/natural"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
	"github.com/sarifsystems/sarif/transports/sfproto"
)

var Module = &services.Module{
	Name:        "xmpp",
	Version:     "1.0",
	NewInstance: New,
}

type Dependencies struct {
	Config        services.Config
	ClientFactory sarif.ClientFactory
	Log           sfproto.Logger
}

type Config struct {
	Server   string
	User     string
	Domain   string
	Password string
}

type conversation struct {
	Remote string
	Proto  sarif.Client
	Xmpp   *Client
}

type Client struct {
	cfg           Config
	Log           sfproto.Logger
	ClientFactory sarif.ClientFactory
	xmpp          *xmpp.Conn
	conversations map[string]*conversation
}

func New(deps *Dependencies) *Client {
	c := &Client{
		Log:           deps.Log,
		ClientFactory: deps.ClientFactory,
		conversations: make(map[string]*conversation, 0),
	}
	deps.Config.Get(&c.cfg)
	return c
}

func (c *Client) Enable() (err error) {
	return c.connectXmpp()
}

func (c *Client) connectXmpp() (err error) {
	cfg := &xmpp.Config{}
	c.xmpp, err = xmpp.Dial(c.cfg.Server, c.cfg.User, c.cfg.Domain, c.cfg.Password, "sarif", cfg)
	if err != nil {
		return err
	}
	go c.listen()
	return c.xmpp.SignalPresence("")
}

func (c *Client) reconnectLoop() {
	for {
		if err := c.connectXmpp(); err != nil {
			c.Log.Debugln("[xmpp] reconnect error:", err)
			time.Sleep(5 * time.Second)
		}
		c.Log.Infoln("[xmpp] reconnected")
		return
	}
}

func (cv *conversation) handleProtoMessage(msg sarif.Message) {
	text := natural.FormatSimple(msg)
	cv.Xmpp.Log.Debugf("[xmpp] send '%s' to '%s'", text, cv.Remote)
	if err := cv.Xmpp.xmpp.Send(cv.Remote, text); err != nil {
		cv.Xmpp.Log.Errorln("[xmpp] send:", err)
	}
}

func (c *Client) listen() {
	for {
		stanza, err := c.xmpp.Next()
		if err != nil {
			c.reconnectLoop()
			return
		}

		switch v := stanza.Value.(type) {
		case *xmpp.ClientMessage:
			c.handleChatMessage(v)
		default:
			c.Log.Debugln("[xmpp] stanza", stanza.Name, v)
		}
	}
}

func (c *Client) newConversation(remote string) *conversation {
	user := xmpp.RemoveResourceFromJid(remote)
	client, _ := c.ClientFactory.NewClient(sarif.ClientInfo{
		Name: "xmpp/" + user,
	})
	cv := &conversation{
		Remote: remote,
		Proto:  client,
		Xmpp:   c,
	}
	if err := client.Subscribe("", "self", cv.handleProtoMessage); err != nil {
		c.Log.Errorln("[xmpp] new:", err)
	}
	c.conversations[user] = cv
	return cv
}

func (c *Client) handleChatMessage(chat *xmpp.ClientMessage) {
	c.Log.Debugln("[xmpp] chat: ", chat)
	if chat.Type != "chat" {
		return
	}
	if chat.Body == "" {
		return
	}

	cv, ok := c.conversations[xmpp.RemoveResourceFromJid(chat.From)]
	if !ok {
		cv = c.newConversation(chat.From)
	}

	if strings.HasPrefix(chat.Body, ".subscribe ") {
		action := strings.TrimPrefix(chat.Body, ".subscribe ")
		if action != "" {
			if err := cv.Proto.Subscribe(action, "", cv.handleProtoMessage); err != nil {
				c.Log.Errorln("[xmpp] subscribe:", err)
			}
		}
		return
	}

	cv.Proto.Publish(sarif.Message{
		Action: "natural/handle",
		Text:   chat.Body,
	})
}
