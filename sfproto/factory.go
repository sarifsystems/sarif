package sfproto

import "github.com/sarifsystems/sarif/sarif"

type protoFactory struct {
	NetConfig NetConfig
}

func (f *protoFactory) NewClient(ci sarif.ClientInfo) (sarif.Client, error) {
	conn, err := Dial(&f.NetConfig)
	if err != nil {
		return nil, err
	}

	c := sarif.NewClient(ci)
	return c, c.Connect(conn)
}

func NewClientFactory(cfg NetConfig) sarif.ClientFactory {
	return &protoFactory{cfg}
}
