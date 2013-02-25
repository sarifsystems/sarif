package terminal

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/service"

	_ "github.com/xconstruct/stark/transport/net"
)

type Terminal struct {
	serv *service.Service
}

func New(url string) *Terminal {
	serv := service.MustConnect(url, service.Info{
		Name: "terminal",
	})
	t := &Terminal{serv}
	return t
}

func (t *Terminal) Handle(msg *stark.Message) (*stark.Message, error) {
	fmt.Println(msg.Message)
	return nil, nil
}

func (t *Terminal) ListenInput() {
	stdin := bufio.NewReader(os.Stdin)

	for {
		cmd, err := stdin.ReadString('\n')
		if err != nil {
			return
		}
		cmd = strings.TrimSpace(cmd)

		msg := stark.NewMessage()
		msg.Action = "natural.process"
		msg.Message = cmd
		t.serv.Write(msg)
	}
}

func (t *Terminal) Start() {
	go t.serv.HandleLoop(t)
	go t.ListenInput()
}
