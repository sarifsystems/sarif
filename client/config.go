package client

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
)

type Config struct {
	DeviceName  string
	DeviceId    string
	Domain      string
	Server      string
	Certificate string
	Key         string
	Authority   string
}

func (cfg *Config) LoadFromEnv() {
	if cfg.Server == "" {
		cfg.Server = os.Getenv("STARK_HOST")
	}
	if cfg.Certificate == "" {
		cfg.Certificate = os.Getenv("STARK_CERT")
	}
	if cfg.Key == "" {
		cfg.Key = os.Getenv("STARK_KEY")
	}
	if cfg.Authority == "" {
		cfg.Authority = os.Getenv("STARK_CA")
	}
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
