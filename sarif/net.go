// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package sarif

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"net/url"
	"strings"
	"time"
)

const DefaultPort = "23100"
const DefaultTlsPort = "23443"
const DefaultKeepalive = 30 * time.Second

type AuthType string

const (
	AuthNone        AuthType = "none"
	AuthChallenge   AuthType = ""
	AuthCertificate          = "certificate"
)

type NetConfig struct {
	Address     string
	Auth        AuthType
	Certificate string
	Key         string
	Authority   string
	Tls         *tls.Config `json:"-"`
	Keepalive   int         `json:",omitempty"`
}

func (cfg *NetConfig) loadTlsCertificates(u *url.URL) error {
	if cfg.Tls != nil {
		return nil
	}

	host := u.Host
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[0:i]
	}

	cfg.Tls = &tls.Config{
		ClientAuth: tls.VerifyClientCertIfGiven,
		ServerName: host,
	}
	if cfg.Certificate != "" && cfg.Key != "" {
		cert, err := tls.LoadX509KeyPair(cfg.Certificate, cfg.Key)
		if err != nil {
			return err
		}
		cfg.Tls.Certificates = []tls.Certificate{cert}
	}

	if cfg.Authority != "" {
		roots := x509.NewCertPool()
		cert, err := ioutil.ReadFile(cfg.Authority)
		if err != nil {
			return err
		}
		roots.AppendCertsFromPEM(cert)
		cfg.Tls.RootCAs = roots
		cfg.Tls.ClientCAs = roots
	}

	return nil
}

func (cfg *NetConfig) parseUrl() (*url.URL, error) {
	u, err := url.Parse(cfg.Address)
	if err != nil {
		return nil, err
	}

	if cfg.Tls == nil {
		if err := cfg.loadTlsCertificates(u); err != nil {
			return nil, err
		}
	}

	if u.Scheme == "" {
		u.Scheme = "tcp"
	}
	if !strings.Contains(u.Host, ":") {
		if cfg.Tls != nil {
			u.Host += ":" + DefaultTlsPort
		} else {
			u.Host += ":" + DefaultPort
		}
	}

	return u, nil
}

// Dial connects to a sarif broker.
func Dial(cfg *NetConfig) (Conn, error) {
	u, err := cfg.parseUrl()
	if err != nil {
		return nil, err
	}

	var conn net.Conn
	if strings.HasSuffix(u.Scheme, "+tls") {
		u.Scheme = strings.TrimSuffix(u.Scheme, "+tls")
		if cfg.Tls == nil {
			return nil, errors.New("expected valid TLS config")
		}
		conn, err = tls.Dial(u.Scheme, u.Host, cfg.Tls)
	} else {
		conn, err = net.Dial(u.Scheme, u.Host)
	}
	if err != nil {
		return nil, err
	}

	ka := time.Duration(cfg.Keepalive) * time.Second
	if ka == 0 {
		ka = DefaultKeepalive
	}
	nc := newNetConn(conn)
	go nc.KeepaliveLoop(ka)
	return nc, nil
}

type NetListener struct {
	cfg *NetConfig
	net.Listener
}

func Listen(cfg *NetConfig) (*NetListener, error) {
	u, err := cfg.parseUrl()
	if err != nil {
		return nil, err
	}

	l := &NetListener{cfg, nil}
	if strings.HasSuffix(u.Scheme, "+tls") {
		u.Scheme = strings.TrimSuffix(u.Scheme, "+tls")
		if cfg.Tls == nil {
			return nil, errors.New("expected valid TLS config")
		}
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
	var conn net.Conn
	if conn, err = l.Listener.Accept(); err != nil {
		return nil, err
	}
	if tc, ok := conn.(*tls.Conn); ok {
		if err := tc.Handshake(); err != nil {
			return nil, err
		}
	}
	return newNetConn(conn), nil
}
