// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"

	mqtt "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"

	"github.com/xconstruct/stark/log"
	"github.com/xconstruct/stark/proto"
)

var (
	ErrNotConnected = errors.New("MQTT transport is not connected")
)

type Config struct {
	Server      string
	Certificate string
	Key         string
	Authority   string
}

func GetDefaults() Config {
	return Config{}
}

func (cfg *Config) LoadTlsCertificates() (*tls.Config, error) {
	tcfg := &tls.Config{}
	if cfg.Certificate != "" && cfg.Key != "" {
		cert, err := tls.LoadX509KeyPair(cfg.Certificate, cfg.Key)
		if err != nil {
			return tcfg, err
		}
		tcfg.Certificates = []tls.Certificate{cert}
	}

	if cfg.Authority != "" {
		roots := x509.NewCertPool()
		cert, err := ioutil.ReadFile(cfg.Authority)
		if err != nil {
			return tcfg, err
		}
		roots.AppendCertsFromPEM(cert)
		tcfg.RootCAs = roots
		tcfg.InsecureSkipVerify = true
	}

	return tcfg, nil
}

type Transport struct {
	client  *mqtt.MqttClient
	cfg     Config
	handler proto.Handler
	log     log.Interface
}

func New(cfg Config) *Transport {
	return &Transport{
		nil,
		cfg,
		nil,
		log.Default,
	}
}

func (t *Transport) Connect() error {
	t.log.Infof("mqtt connecting to %s", t.cfg.Server)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(t.cfg.Server)
	opts.SetClientId(proto.GenerateId())
	opts.SetCleanSession(true)
	opts.SetOnConnectionLost(t.onConnectionLost)
	tlscfg, err := t.cfg.LoadTlsCertificates()
	if err != nil {
		return err
	}
	opts.SetTlsConfig(tlscfg)
	t.client = mqtt.NewClient(opts)
	if _, err := t.client.Start(); err != nil {
		return err
	}

	return nil
}

func (t *Transport) IsConnected() bool {
	return t.client != nil && t.client.IsConnected()
}

func (t *Transport) Publish(msg proto.Message) error {
	if !t.IsConnected() {
		return ErrNotConnected
	}
	if msg.Action == "proto/sub" {
		action := msg.PayloadGetString("action")
		device := msg.PayloadGetString("device")
		if err := t.subscribe(proto.GetTopic(action, device) + "/#"); err != nil {
			return err
		}
	}

	raw, err := msg.Encode()
	if err != nil {
		return err
	}

	topic := proto.GetTopic(msg.Action, msg.Destination)
	t.log.Debugf("mqtt sending to %s: %v", topic, string(raw))
	r := t.client.Publish(mqtt.QOS_ZERO, topic, raw)
	<-r
	return nil
}

func (t *Transport) subscribe(topic string) error {
	if !t.IsConnected() {
		return ErrNotConnected
	}
	t.log.Debugln("mqtt subscribing to", topic)
	filter, err := mqtt.NewTopicFilter(topic, 0)
	if err != nil {
		return err
	}
	if _, err := t.client.StartSubscription(t.handleRawMessage, filter); err != nil {
		return err
	}
	return nil
}

func (t *Transport) RegisterHandler(h proto.Handler) {
	t.handler = h
}

func (t *Transport) handleRawMessage(client *mqtt.MqttClient, raw mqtt.Message) {
	m, err := proto.DecodeMessage(raw.Payload())
	if err != nil {
		t.log.Warnln(err)
		return
	}
	t.log.Debugf("mqtt receiving from %s: %v", raw.Topic(), string(raw.Payload()))
	t.handler(m)
}

func (t *Transport) onConnectionLost(client *mqtt.MqttClient, reason error) {
	t.log.Infoln("mqtt transport lost connection:", reason)
}
