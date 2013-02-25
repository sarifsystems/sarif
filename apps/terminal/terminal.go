package main

import (
	"github.com/xconstruct/stark/service/terminal"

	_ "github.com/xconstruct/stark/transport/net"
)

func main() {
	t := terminal.New("tcp://127.0.0.1")
	t.Start()

	switch {}
}
