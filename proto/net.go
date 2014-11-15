// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package proto

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"strings"
)

var DefaultPort = "23100"

type NetConfig struct {
	Address     string
	Certificate string
	Key         string
	Authority   string
	Tls         *tls.Config `json:"-"`
}

func (cfg *NetConfig) loadTlsCertificates() error {
	if cfg.Certificate == "" || cfg.Key == "" {
		return nil
	}

	cfg.Tls = &tls.Config{}
	cert, err := tls.LoadX509KeyPair(cfg.Certificate, cfg.Key)
	if err != nil {
		return err
	}
	cfg.Tls.Certificates = []tls.Certificate{cert}

	if cfg.Authority != "" {
		roots := x509.NewCertPool()
		cert, err := ioutil.ReadFile(cfg.Authority)
		if err != nil {
			return err
		}
		roots.AppendCertsFromPEM(cert)
		cfg.Tls.RootCAs = roots
		cfg.Tls.InsecureSkipVerify = true
	}

	return nil
}

func (cfg *NetConfig) parseUrl() (*url.URL, error) {
	u, err := url.Parse(cfg.Address)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		u.Scheme = "tcp"
	}
	if !strings.Contains(u.Host, ":") {
		u.Host += ":" + DefaultPort
	}

	return u, nil
}

// Dial connects to a stark broker.
func Dial(cfg *NetConfig) (Conn, error) {
	u, err := cfg.parseUrl()
	if err != nil {
		return nil, err
	}

	var conn io.ReadWriteCloser
	if cfg.Tls != nil {
		conn, err = tls.Dial(u.Scheme, u.Host, cfg.Tls)
	} else {
		conn, err = net.Dial(u.Scheme, u.Host)
	}
	if err != nil {
		return nil, err
	}
	return NewByteConn(conn), nil
}

type NetListener struct {
	net.Listener
}

func Listen(cfg *NetConfig) (*NetListener, error) {
	u, err := cfg.parseUrl()
	if err != nil {
		return nil, err
	}

	if cfg.Tls == nil {
		if err := cfg.loadTlsCertificates(); err != nil {
			return nil, err
		}
	}

	l := &NetListener{}
	if cfg.Tls != nil {
		l.Listener, err = tls.Listen(u.Scheme, u.Host, cfg.Tls)
	} else {
		l.Listener, err = net.Listen(u.Scheme, u.Host)
	}
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (l *NetListener) Accept() (Conn, error) {
	var err error
	var conn io.ReadWriteCloser
	if conn, err = l.Listener.Accept(); err != nil {
		return nil, err
	}
	return NewByteConn(conn), nil
}
