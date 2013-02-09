package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/xconstruct/stark"
	"github.com/xconstruct/stark/service"

	_ "github.com/xconstruct/stark/transport/net"
)

func main() {
	s, err := service.Connect("tcp://127.0.0.1", service.Info{
		Name: "terminal",
	})
	if err != nil {
		log.Fatalf("terminal: %v", err)
	}
	go func() {
		stdin := bufio.NewReader(os.Stdin)
		for {
			cmd, _ := stdin.ReadString('\n')
			cmd = strings.TrimSpace(cmd)

			msg := stark.NewMessage()
			msg.Action = "natural.process"
			msg.Destination = "natural"
			msg.Message = cmd
			s.Write(msg)
		}
	}()

	for {
		msg, err := s.Read()
		if err != nil {
			return
		}
		fmt.Println(msg.Message)
	}
}
