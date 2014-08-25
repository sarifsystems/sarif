package client

import (
	"github.com/xconstruct/stark/log"
	"github.com/xconstruct/stark/proto"
)

type Client struct {
	DeviceName string
	DeviceId   string
	endpoint   proto.Endpoint
	handler    proto.Handler
}

func New(deviceName string, e proto.Endpoint) *Client {
	c := &Client{
		deviceName,
		deviceName + "-" + proto.GenerateId(),
		e,
		nil,
	}
	c.SetEndpoint(e)
	return c
}

func (c *Client) SetEndpoint(e proto.Endpoint) {
	c.endpoint = e
	e.RegisterHandler(c.handle)
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
	c.FillMessage(&msg)
	return c.endpoint.Publish(msg)
}

func (c *Client) handle(msg proto.Message) {
	if msg.Action == "ping" {
		c.handlePing(msg)
	}
	c.handler(msg)
}

func (c *Client) handlePing(msg proto.Message) {
	log.Default.Debugf("[client] %s got ping", c.DeviceId)
	err := c.Publish(msg.Reply(proto.Message{
		Action: "ack",
		CorrId: msg.Id,
	}))
	if err != nil {
		log.Default.Warnln(err)
	}
}

func (c *Client) RegisterHandler(h proto.Handler) {
	c.handler = h
	c.SubscribeGlobal("ping")
}

func (c *Client) SubscribeGlobal(action string) error {
	if err := c.SubscribeSelf(action); err != nil {
		return err
	}
	return c.Publish(proto.Message{
		Action: "proto/sub",
		Payload: map[string]interface{}{
			"action": action,
			"device": "",
		},
	})
}

func (c *Client) SubscribeSelf(action string) error {
	return c.Publish(proto.Message{
		Action: "proto/sub",
		Payload: map[string]interface{}{
			"action": action,
			"device": c.DeviceId,
		},
	})
}
