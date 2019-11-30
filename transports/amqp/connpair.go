// Copyright (C) 2019 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package amqp

import (
	"github.com/sarifsystems/sarif/transports/sfproto"
)

type connpair struct {
	in  *amqpConn
	out *amqpConn
}

func internalDial(cfg *sfproto.NetConfig) (*connpair, error) {
	in := &amqpConn{cfg: cfg}
	if err := in.Dial(); err != nil {
		return nil, err
	}

	out := &amqpConn{cfg: cfg}
	if err := out.Dial(); err != nil {
		return nil, err
	}

	return &connpair{in, out}, nil
}
