package main

import (
	"github.com/xconstruct/stark/core"

	_ "github.com/xconstruct/stark/services/hostscan"
	_ "github.com/xconstruct/stark/services/router"
)

type Config struct {
	EnabledModules []string
}

func main() {
	ctx, err := core.NewContext("stark")
	ctx.Must(err)
	defer ctx.Close()

	var cfg Config
	ctx.Must(ctx.Config.Get("server", &cfg))

	for _, module := range cfg.EnabledModules {
		ctx.Must(ctx.EnableModule(module))
	}

	ctx.WaitUntilInterrupt()
}
