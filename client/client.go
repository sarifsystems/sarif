package client

import (
	_ "crypto/md5"
	"strings"

	mqtt "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"

	"github.com/xconstruct/stark/log"
)

const VERSION = "0.2.0"

type Client struct {
	cfg           Config
	client        *mqtt.MqttClient
	subscriptions []subscription
}

type MessageHandler func(msg Message)

type subscription struct {
	Action  string
	Handler MessageHandler
}

func New(cfg Config) *Client {
	if cfg.DeviceId == "" {
		cfg.DeviceId = cfg.DeviceName + "-" + GenerateId()
	}
	c := &Client{cfg, nil, make([]subscription, 0)}
	return c
}

func (c *Client) Connect() error {
	log.Default.Debugf("connecting to %s", c.cfg.Server)
	opts := mqtt.NewClientOptions()
	opts.SetBroker(c.cfg.Server)
	opts.SetClientId(c.cfg.DeviceId)
	opts.SetCleanSession(true)
	opts.SetTlsConfig(c.cfg.TlsConfig)
	opts.SetTraceLevel(mqtt.Critical)

	c.client = mqtt.NewClient(opts)
	if _, err := c.client.Start(); err != nil {
		return err
	}

	filter, err := mqtt.NewTopicFilter(GetTopic("", c.cfg.DeviceId, ""), 0)
	if err != nil {
		return err
	}
	if _, err := c.client.StartSubscription(c.handleRawMessage, filter); err != nil {
		return err
	}
	if err = c.Subscribe("ping", c.handlePing); err != nil {
		return err
	}

	return nil
}

func (c *Client) FillMessage(msg *Message) {
	if msg.Version == "" {
		msg.Version = VERSION
	}
	if msg.Id == "" {
		msg.Id = GenerateId()
	}
	if msg.Source == "" {
		msg.Source = c.cfg.DeviceId
	}
}

func (c *Client) Publish(msg Message) error {
	c.FillMessage(&msg)
	raw, err := msg.Encode()
	if err != nil {
		return err
	}
	log.Default.Debugln("sending message", string(raw))
	topic := GetTopic(msg.Action, msg.Device, msg.Domain)
	r := c.client.Publish(mqtt.QOS_ZERO, topic, raw)
	<-r
	return nil
}

func (c *Client) handleRawMessage(client *mqtt.MqttClient, raw mqtt.Message) {
	m, err := DecodeMessage(raw.Payload())
	if err != nil {
		log.Default.Warnln(err)
		return
	}
	log.Default.Debugln("receiving message", string(raw.Payload()))

	for _, sub := range c.subscriptions {
		if strings.HasPrefix(m.Action+"/", sub.Action+"/") {
			sub.Handler(m)
		}
	}
}

func (c *Client) handlePing(msg Message) {
	err := c.Publish(msg.Reply(Message{
		Action: "ack",
	}))
	if err != nil {
		log.Default.Warnln(err)
	}
}

func (c *Client) Subscribe(action string, handler MessageHandler) error {
	c.subscriptions = append(c.subscriptions, subscription{action, handler})
	filter, err := mqtt.NewTopicFilter(GetTopic(action, "", ""), 0)
	if err != nil {
		return err
	}
	log.Default.Debugf("[client] subscribing to '%s'", action)
	_, err = c.client.StartSubscription(c.handleRawMessage, filter)
	return err
}
