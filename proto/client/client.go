package client

import (
	"github.com/xconstruct/stark/log"
	"github.com/xconstruct/stark/proto"
	"strings"
)

type Client struct {
	DeviceName    string
	DeviceId      string
	subscriptions []subscription
	transport     Transport
	connected     bool
}

type Transport interface {
	Connect(deviceId string, msg proto.MessageHandler) error
	Publish(msg proto.Message) error
	Subscribe(action, device, domain string) error
}

type subscription struct {
	Action  string
	Handler proto.MessageHandler
}

func New(deviceName string) *Client {
	c := &Client{
		deviceName,
		deviceName + "-" + proto.GenerateId(),
		make([]subscription, 0),
		nil,
		false,
	}
	return c
}

func (c *Client) SetTransport(t Transport) {
	c.transport = t
}

func (c *Client) Connect() error {
	if err := c.transport.Connect(c.DeviceId, c.handleMessage); err != nil {
		return err
	}
	c.connected = true

	if err := c.Subscribe("ping", c.handlePing); err != nil {
		return err
	}

	if err := c.transport.Subscribe("", c.DeviceName, ""); err != nil {
		return err
	}
	if err := c.transport.Subscribe("", c.DeviceId, ""); err != nil {
		return err
	}
	return nil
}

func (c *Client) FillMessage(msg *proto.Message) {
	if msg.Version == "" {
		msg.Version = proto.VERSION
	}
	if msg.Id == "" {
		msg.Id = proto.GenerateId()
	}
	if msg.Source == "" {
		msg.Source = c.DeviceId
	}
}

func (c *Client) Publish(msg proto.Message) error {
	if !c.connected {
		err := c.Connect()
		if err != nil {
			return err
		}
	}
	c.FillMessage(&msg)
	return c.transport.Publish(msg)
}

func (c *Client) handleMessage(msg proto.Message) {
	for _, sub := range c.subscriptions {
		if strings.HasPrefix(msg.Action+"/", sub.Action+"/") {
			go sub.Handler(msg)
		}
	}
}

func (c *Client) handlePing(msg proto.Message) {
	err := c.Publish(msg.Reply(proto.Message{
		Action: "ack",
	}))
	if err != nil {
		log.Default.Warnln(err)
	}
}

func (c *Client) Subscribe(action string, handler proto.MessageHandler) error {
	if !c.connected {
		err := c.Connect()
		if err != nil {
			return err
		}
	}
	log.Default.Debugf("[client] subscribing to '%s'", action)
	c.subscriptions = append(c.subscriptions, subscription{action, handler})
	return c.transport.Subscribe(action, "", "")
}
