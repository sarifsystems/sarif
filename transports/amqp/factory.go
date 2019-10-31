package amqp

import (
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/transports/sfproto"
)

type amqpFactory struct {
	NetConfig  sfproto.NetConfig
	Connection *connpair
}

func (f *amqpFactory) NewClient(ci sarif.ClientInfo) (sarif.Client, error) {
	if f.Connection == nil {
		aconn, err := internalDial(&f.NetConfig)
		if err != nil {
			return nil, err
		}
		f.Connection = aconn
	}

	conn, err := newConnection(f.Connection)
	if err != nil {
		return nil, err
	}

	c := sarif.NewClient(ci)
	return c, c.Connect(conn)
}

func NewClientFactory(cfg sfproto.NetConfig) sarif.ClientFactory {
	return &amqpFactory{NetConfig: cfg}
}
