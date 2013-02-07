package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/xconstruct/stark/natural"
	"github.com/xconstruct/stark/transports/net"
)

func main() {
	conn, err := net.Connect("tcp", "127.0.0.1:9000")
	if err != nil {
		log.Fatalf("terminal: %v", err)
	}
	go func() {
		stdin := bufio.NewReader(os.Stdin)
		for {
			cmd, _ := stdin.ReadString('\n')
			cmd = strings.TrimSpace(cmd)

			msg := natural.NewMessage("temp", cmd)
			conn.Write(msg)
		}
	}()

	for {
		msg, err := conn.Read()
		if err != nil {
			return
		}
		fmt.Println(msg.Message)
	}
}
