package sfproto

import (
	"github.com/sarifsystems/sarif/sarif"
)

type wrappedConn struct {
	id string
	Conn
}

func (c *wrappedConn) Publish(msg sarif.Message) error {
	return c.Write(msg)
}

func (c *wrappedConn) Subscribe(src, action, dest string) error {
	msg := Subscribe(action, dest)
	msg.Source = src
	return c.Publish(msg)
}

func (c *wrappedConn) Consume() (<-chan sarif.Message, error) {
	ch := make(chan sarif.Message, 10)

	go func() {
		for {
			msg, err := c.Read()
			if err != nil {
				close(ch)
				c.Close()
				c.Conn = nil
				return
			}
			ch <- msg
		}
	}()

	return (<-chan sarif.Message)(ch), nil
}

func wrap(conn Conn) sarif.Connection {
	return &wrappedConn{sarif.GenerateId(), conn}
}
