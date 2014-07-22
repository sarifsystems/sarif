package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	mqtt "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"

	"github.com/xconstruct/stark/log"
	"github.com/xconstruct/stark/proto"
)

type Config struct {
	Server      string
	Certificate string
	Key         string
	Authority   string
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
	handler proto.MessageHandler
}

func New(cfg Config) *Transport {
	return &Transport{nil, cfg, nil}
}

func (t *Transport) Connect(deviceId string, handler proto.MessageHandler) error {
	log.Default.Debugf("[mqtt] connecting to %s", t.cfg.Server)
	t.handler = handler

	opts := mqtt.NewClientOptions()
	opts.SetBroker(t.cfg.Server)
	opts.SetClientId(deviceId)
	opts.SetCleanSession(true)
	opts.SetTraceLevel(mqtt.Critical)
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

func (t *Transport) Publish(msg proto.Message) error {
	raw, err := msg.Encode()
	if err != nil {
		return err
	}
	log.Default.Debugln("[mqtt] sending message", string(raw))
	topic := proto.GetTopic(msg.Action, msg.Device, msg.Domain)
	r := t.client.Publish(mqtt.QOS_ZERO, topic, raw)
	<-r
	return nil
}

func (t *Transport) handleRawMessage(client *mqtt.MqttClient, raw mqtt.Message) {
	m, err := proto.DecodeMessage(raw.Payload())
	if err != nil {
		log.Default.Warnln(err)
		return
	}
	log.Default.Debugln("receiving message", string(raw.Payload()))
	t.handler(m)
}

func (t *Transport) Subscribe(action, device, domain string) error {
	log.Default.Debugf("[client] subscribing to %s@%s/%s", action, domain, device)
	filter, err := mqtt.NewTopicFilter(proto.GetTopic(action, device, domain), 0)
	if err != nil {
		return err
	}
	if _, err := t.client.StartSubscription(t.handleRawMessage, filter); err != nil {
		return err
	}

	return nil
}
