// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package xmpp

import (
	"encoding/json"
	"time"

	"github.com/agl/xmpp"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/proto/natural"
)

var Module = core.Module{
	Name:        "xmpp",
	Version:     "1.0",
	NewInstance: NewInstance,
}

func init() {
	core.RegisterModule(Module)
}

type Config struct {
	Server   string
	User     string
	Domain   string
	Password string
}

type conversation struct {
	Remote      string
	Proto       *proto.Client
	LastMessage proto.Message
	Xmpp        *Client
}

type Client struct {
	cfg           Config
	ctx           *core.Context
	mux           *proto.Mux
	xmpp          *xmpp.Conn
	conversations map[string]*conversation
}

func New(ctx *core.Context) (*Client, error) {
	c := &Client{
		ctx:           ctx,
		conversations: make(map[string]*conversation, 0),
	}
	err := ctx.Config.Get("xmpp", &c.cfg)
	return c, err
}

func NewInstance(ctx *core.Context) (core.ModuleInstance, error) {
	return New(ctx)
}

func (c *Client) Enable() (err error) {
	c.mux = proto.NewMux()
	proto.Connect(c.ctx.Proto, c.mux)
	return c.connectXmpp()
}

func (c *Client) connectXmpp() (err error) {
	cfg := &xmpp.Config{}
	c.xmpp, err = xmpp.Dial(c.cfg.Server, c.cfg.User, c.cfg.Domain, c.cfg.Password, cfg)
	if err != nil {
		return err
	}
	go c.listen()
	return c.xmpp.SignalPresence("")
}

func (c *Client) reconnectLoop() {
	for {
		if err := c.connectXmpp(); err != nil {
			c.ctx.Log.Debugln("[xmpp] reconnect error:", err)
			time.Sleep(5 * time.Second)
		}
		c.ctx.Log.Infoln("[xmpp] reconnected")
		return
	}
}

func (c *Client) Disable() error {
	return nil
}

func (cv *conversation) handleProtoMessage(msg proto.Message) {
	cv.LastMessage = msg
	text := natural.FormatSimple(msg)
	cv.Xmpp.ctx.Log.Debugf("[xmpp] send '%s' to '%s'", text, cv.Remote)
	if err := cv.Xmpp.xmpp.Send(cv.Remote, text); err != nil {
		cv.Xmpp.ctx.Log.Errorln("[xmpp] send:", err)
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
			c.ctx.Log.Debugln("[xmpp] stanza", stanza.Name, v)
		}
	}
}

func (c *Client) newConversation(remote string) *conversation {
	ep := c.mux.NewEndpoint()
	client := proto.NewClient("xmpp-"+proto.GenerateId(), ep)
	cv := &conversation{
		Remote: remote,
		Proto:  client,
		Xmpp:   c,
	}
	if err := client.Subscribe("", "self", cv.handleProtoMessage); err != nil {
		c.ctx.Log.Errorln("[xmpp] new:", err)
	}
	c.conversations[cv.Remote] = cv
	return cv
}

func (c *Client) handleChatMessage(chat *xmpp.ClientMessage) {
	c.ctx.Log.Debugln("[xmpp] chat: ", chat)
	if chat.Body == "" {
		return
	}

	cv, ok := c.conversations[chat.From]
	if !ok {
		cv = c.newConversation(chat.From)
	}

	if chat.Body == "full" {
		text, err := json.MarshalIndent(cv.LastMessage, "", "    ")
		if err != nil {
			panic(err)
		}
		if err := c.xmpp.Send(chat.From, string(text)); err != nil {
			c.ctx.Log.Errorln("[xmpp] send:", err)
		}
		return
	}

	cv.Proto.Publish(proto.Message{
		Action: "natural/handle",
		Text:   chat.Body,
	})
}
