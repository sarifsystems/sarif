package main

import (
	"github.com/xconstruct/stark/client"
	"github.com/xconstruct/stark/log"
)

func assert(err error) {
	if err != nil {
		log.Default.Fatalln(err)
	}
}

func main() {
	cfg := client.Config{
		DeviceName: "starkping",
	}
	assert(cfg.LoadFromEnv())

	c := client.New(cfg)
	assert(c.Connect())

	err := c.Publish(client.Message{
		Action: "ping",
	})
	select {}
	assert(err)
}
