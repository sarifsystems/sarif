package main

import (
	"flag"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/log"

	_ "github.com/xconstruct/stark/services/hostscan"
	_ "github.com/xconstruct/stark/services/router"
	_ "github.com/xconstruct/stark/web"
)

var verbose = flag.Bool("v", false, "verbose debug output")

type Config struct {
	EnabledModules []string
}

func main() {
	flag.Parse()

	ctx, err := core.NewContext("stark")
	ctx.Must(err)
	defer ctx.Close()

	if *verbose {
		ctx.Log.SetLevel(log.LevelDebug)
	}

	cfg := Config{
		EnabledModules: []string{"web"},
	}
	ctx.Must(ctx.Config.Get("server", &cfg))

	for _, module := range cfg.EnabledModules {
		ctx.Must(ctx.EnableModule(module))
	}

	ctx.WaitUntilInterrupt()
}
