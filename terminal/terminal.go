package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/service"

	_ "github.com/xconstruct/stark/transport/net"
)

func main() {
	s := service.MustConnect("tcp://127.0.0.1", service.Info{
		Name: "terminal",
	})

	go func() {
		for {
			msg, err := s.Read()
			if err != nil {
				return
			}
			fmt.Println(msg.Message)
		}
	}()

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
		s.Write(msg)
	}
}
