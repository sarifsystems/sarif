package xmpp

import (
	"fmt"

	"github.com/agl/xmpp"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/proto/client"
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
	MessageId string
	Remote    string
}

type Client struct {
	cfg           Config
	ctx           *core.Context
	proto         *client.Client
	xmpp          *xmpp.Conn
	conversations []conversation
	lastMessage   proto.Message
}

func New(ctx *core.Context) (*Client, error) {
	c := &Client{
		ctx:           ctx,
		conversations: make([]conversation, 0),
	}
	err := ctx.Config.Get("xmpp", &c.cfg)
	return c, err
}

func NewInstance(ctx *core.Context) (core.ModuleInstance, error) {
	return New(ctx)
}

func (c *Client) Enable() (err error) {
	cfg := &xmpp.Config{}
	c.xmpp, err = xmpp.Dial(c.cfg.Server, c.cfg.User, c.cfg.Domain, c.cfg.Password, cfg)
	if err != nil {
		return err
	}

	c.proto = c.ctx.NewProtoClient("xmpp")
	c.proto.RegisterHandler(c.handleProtoMessage)
	if err := c.proto.SubscribeSelf(""); err != nil {
		return err
	}

	go c.listen()
	return c.xmpp.SignalPresence("")
}

func (c *Client) Disable() error {
	return nil
}

func (c *Client) handleProtoMessage(msg proto.Message) {
	if msg.CorrId == "" {
		c.ctx.Log.Debugln("[xmpp] received proto msg: ", msg)
		return
	}

	var cv *conversation
	for _, v := range c.conversations {
		if v.MessageId == msg.CorrId {
			cv = &v
			break
		}
	}
	if cv == nil {
		c.ctx.Log.Debugln("[xmpp] received proto msg: ", msg)
	}

	c.lastMessage = msg
	text := natural.FormatSimple(msg)
	c.ctx.Log.Debugf("[xmpp] send '%s' to '%s'", text, cv.Remote)
	if err := c.xmpp.Send(cv.Remote, text); err != nil {
		c.ctx.Log.Errorln("[xmpp] send:", err)
	}
}

func (c *Client) listen() {
	for {
		stanza, err := c.xmpp.Next()
		if err != nil {
			c.ctx.Log.Fatalln("[xmpp]", err)
		}

		switch v := stanza.Value.(type) {
		case *xmpp.ClientMessage:
			c.handleChatMessage(v)
		default:
			c.ctx.Log.Debugln("[xmpp] stanza", stanza.Name, v)
		}
	}
}

func (c *Client) handleChatMessage(chat *xmpp.ClientMessage) {
	c.ctx.Log.Debugln("[xmpp] chat: ", chat)
	if chat.Body == "" {
		return
	}

	if chat.Body == "full" {
		text := fmt.Sprintf("%v", c.lastMessage)
		if err := c.xmpp.Send(chat.From, text); err != nil {
			c.ctx.Log.Errorln("[xmpp] send:", err)
		}
		return
	}

	msg, ok := natural.ParseSimple(chat.Body)
	if !ok {
		if err := c.xmpp.Send(chat.From, "I didn't understand your message."); err != nil {
			c.ctx.Log.Errorln("[xmpp] send:", err)
		}
		return
	}

	msg.Id = proto.GenerateId()
	c.conversations = append(c.conversations, conversation{
		MessageId: msg.Id,
		Remote:    chat.From,
	})

	if err := c.proto.Publish(msg); err != nil {
		c.ctx.Log.Errorln("[xmpp] publish: ", err)
	}
}