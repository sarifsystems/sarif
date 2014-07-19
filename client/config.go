package client

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"
)

type Config struct {
	DeviceName string
	DeviceId   string
	Domain     string
	Server     string
	TlsConfig  *tls.Config
}

func (cfg *Config) LoadFromEnv() error {
	if cfg.Server == "" {
		cfg.Server = os.Getenv("STARK_HOST")
	}
	if cfg.TlsConfig == nil {
		cert := os.Getenv("STARK_CERT")
		key := os.Getenv("STARK_KEY")
		ca := os.Getenv("STARK_CA")
		tcfg, err := LoadTlsCertificates(cert, key, ca)
		if err != nil {
			return err
		}
		cfg.TlsConfig = tcfg
	}
	return nil
}

func LoadTlsCertificates(certFile, keyFile, caFile string) (*tls.Config, error) {
	tcfg := &tls.Config{}
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return tcfg, err
		}
		tcfg.Certificates = []tls.Certificate{cert}
	}

	if caFile != "" {
		roots := x509.NewCertPool()
		cert, err := ioutil.ReadFile(caFile)
		if err != nil {
			return tcfg, err
		}
		roots.AppendCertsFromPEM(cert)
		tcfg.RootCAs = roots
		tcfg.InsecureSkipVerify = true
	}

	return tcfg, nil
}
