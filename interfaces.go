package stark

type Conn interface {
	Read() (*Message, error)
	Write(*Message) error
	Close() error
}

func StartReading(conn Conn) (<-chan interface{}) {
	c := make(chan interface{})
	go func() {
		for {
			msg, err := conn.Read()
			if err != nil {
				c <-  err
				return
			} else {
				c <- msg
			}
		}
	}()
	return c
}
