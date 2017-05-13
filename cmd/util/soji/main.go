// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// A client for presenting a statusbar in the shell/tmux.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/sarifsystems/sarif/core"
)

var isTmux = flag.Bool("tmux", false, "tmux status line output")
var hideLabels = flag.Bool("h", false, "hide labels")
var delimiter = flag.String("d", " ", "delimiter")

func main() {
	flag.Parse()
	color.NoColor = false

	app := &app{}
	app.App = core.NewApp("sarif", "tars")
	app.Init()

	app.Info = Cache{
		FilePath: app.Config.Dir() + "/soji.json",
		Bars:     make(map[string]*Bar),
	}

	if flag.Arg(0) == "relay" {
		r := Relay{app: app}
		r.Relay()
		return
	}

	w := Watcher{app}
	w.Watch()
}

type app struct {
	*core.App
	Info Cache
}

type Bar struct {
	Key     string      `json:"name"`
	Label   string      `json:"key"`
	Display string      `json:"display"`
	Value   interface{} `json:"value"`
	Color   string      `json:"color"`
}

func (b Bar) PrettyString(mode string) string {
	lbl := b.Label
	if lbl == "hidden" || *hideLabels {
		lbl = ""
	} else {
		lbl = b.Key
	}
	if lbl != "" {
		lbl += " "
	}

	if b.Value == nil {
		return ""
	}

	text, color := "", "green"
	switch v := b.Value.(type) {
	case string:
		text = v
	case bool:
		if v {
			text, color = "✓", "green"
		} else {
			text, color = "✗", "red"
		}
	default:
		text = fmt.Sprintf("%v", v)
	}
	if text == "" {
		return ""
	}
	if b.Color != "" {
		color = b.Color
	}

	return colorize(mode, "hi-white", lbl) + colorize(mode, color, text)
}

type Cache struct {
	FilePath string `json:"-"`

	DeviceId      string `json:"device_id"`
	ActionChanged string `json:"action_changed"`
	ActionPull    string `json:"action_pull"`

	PID         int       `json:"pid,omitempty"`
	LastUpdated time.Time `json:"time,omitempty"`

	Bars map[string]*Bar `json:"bars,omitempty"`
}

func (c *Cache) Read() error {
	f, err := os.Open(c.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(c)
}

func (c *Cache) Write() error {
	c.LastUpdated = time.Now()

	f, err := os.Create(c.FilePath)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	return err
}

func (c *Cache) Bar(key string) *Bar {
	bar, ok := c.Bars[key]
	if !ok {
		bar = &Bar{Key: key}
		c.Bars[key] = bar
	}
	return bar
}
