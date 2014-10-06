// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"time"

	mqtt "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"

	"github.com/xconstruct/stark/log"
)

var (
	ErrNotConnected = errors.New("MQTT transport is not connected")
)

type MqttConfig struct {
	Server      string
	Certificate string
	Key         string
	Authority   string
	TlsConfig   *tls.Config `json:"-"`
}

func GetMqttDefaults() MqttConfig {
	return MqttConfig{
		Server: "tcp://example.org:1883",
	}
}

func (cfg *MqttConfig) loadTlsCertificates() error {
	cfg.TlsConfig = &tls.Config{}
	if cfg.Certificate != "" && cfg.Key != "" {
		cert, err := tls.LoadX509KeyPair(cfg.Certificate, cfg.Key)
		if err != nil {
			return err
		}
		cfg.TlsConfig.Certificates = []tls.Certificate{cert}
	}

	if cfg.Authority != "" {
		roots := x509.NewCertPool()
		cert, err := ioutil.ReadFile(cfg.Authority)
		if err != nil {
			return err
		}
		roots.AppendCertsFromPEM(cert)
		cfg.TlsConfig.RootCAs = roots
		cfg.TlsConfig.InsecureSkipVerify = true
	}

	return nil
}

type MqttConn struct {
	client        *mqtt.MqttClient
	cfg           MqttConfig
	handler       Handler
	log           log.Interface
	subscriptions map[string]struct{}
}

func DialMqtt(cfg MqttConfig) *MqttConn {
	return &MqttConn{
		nil,
		cfg,
		nil,
		log.Default,
		make(map[string]struct{}),
	}
}

func (t *MqttConn) Connect() error {
	t.log.Infof("mqtt connecting to %s", t.cfg.Server)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(t.cfg.Server)
	opts.SetClientId(GenerateId())
	opts.SetCleanSession(true)
	opts.SetOnConnectionLost(t.onConnectionLost)
	if err := t.cfg.loadTlsCertificates(); err != nil {
		return err
	}
	opts.SetTlsConfig(t.cfg.TlsConfig)
	t.client = mqtt.NewClient(opts)
	if _, err := t.client.Start(); err != nil {
		return err
	}

	return nil
}

func (t *MqttConn) IsConnected() bool {
	return t.client != nil && t.client.IsConnected()
}

func (t *MqttConn) Publish(msg Message) error {
	if !t.IsConnected() {
		return ErrNotConnected
	}
	if msg.Action == "proto/sub" {
		sub := Subscription{}
		if err := msg.DecodePayload(&sub); err != nil {
			return err
		}
		if err := t.subscribe(getTopic(sub.Action, sub.Device) + "/#"); err != nil {
			return err
		}
	}

	raw, err := msg.Encode()
	if err != nil {
		return err
	}

	topic := getTopic(msg.Action, msg.Destination)
	t.log.Debugf("mqtt sending to %s: %v", topic, string(raw))
	r := t.client.Publish(mqtt.QOS_ZERO, topic, raw)
	<-r
	return nil
}

func (t *MqttConn) subscribe(topic string) error {
	if !t.IsConnected() {
		return ErrNotConnected
	}
	t.log.Debugln("mqtt subscribing to", topic)
	filter, err := mqtt.NewTopicFilter(topic, 0)
	if err != nil {
		return err
	}
	t.subscriptions[topic] = struct{}{}
	if _, err := t.client.StartSubscription(t.handleRawMessage, filter); err != nil {
		return err
	}
	return nil
}

func (t *MqttConn) RegisterHandler(h Handler) {
	t.handler = h
}

func (t *MqttConn) handleRawMessage(client *mqtt.MqttClient, raw mqtt.Message) {
	m, err := DecodeMessage(raw.Payload())
	if err != nil {
		t.log.Warnln(err)
		return
	}
	t.log.Debugf("mqtt receiving from %s: %v", raw.Topic(), string(raw.Payload()))
	t.handler(m)
}

func (t *MqttConn) reconnectLoop() {
RECONNECT:
	for {
		if err := t.Connect(); err != nil {
			t.log.Debugln("mqtt reconnect error:", err)
			time.Sleep(5 * time.Second)
		}
		t.log.Infoln("mqtt reconnected")
		for topic := range t.subscriptions {
			if err := t.subscribe(topic); err != nil {
				t.log.Debugln("mqtt reconnect subscribe error:", err)
				time.Sleep(5 * time.Second)
				continue RECONNECT
			}
		}
		return
	}
}

func (t *MqttConn) onConnectionLost(client *mqtt.MqttClient, reason error) {
	t.log.Infoln("mqtt transport lost connection:", reason)
	t.reconnectLoop()
}
