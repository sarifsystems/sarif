package main

import (
	"flag"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/log"
	"github.com/xconstruct/stark/web"

	_ "github.com/xconstruct/stark/services/hostscan"
	_ "github.com/xconstruct/stark/services/router"
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

	var cfg Config
	ctx.Must(ctx.Config.Get("server", &cfg))

	for _, module := range cfg.EnabledModules {
		ctx.Must(ctx.EnableModule(module))
	}

	w, err := web.New(ctx)
	ctx.Must(err)
	ctx.Must(w.Start())

	ctx.WaitUntilInterrupt()
}
