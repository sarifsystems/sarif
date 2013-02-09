package terminal

import (
	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/transports/net"
)

type Terminal struct {
	conn stark.Conn
}

func NewTerminal() *Terminal {
}

func (t *Terminal) Start(conn stark.Conn) error {
	t.conn = conn
	if err != nil {
		log.Fatalf("terminal: %v", err)
	}
	go func() {
		stdin := bufio.NewReader(os.Stdin)
		for {
			cmd, _ := stdin.ReadString('\n')
			cmd = strings.TrimSpace(cmd)

			msg := natural.NewMessage("temp", cmd)
			err := t.conn.Write(msg)
		}
	}()

	for {
		msg, err := t.conn.Read()
		if err != nil {
			return
		}
		fmt.Println(msg.Message)
	}
}

func (t *Terminal( Stop() error {
	t.conn.Close()
}
