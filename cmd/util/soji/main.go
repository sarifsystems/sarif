// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// A client for presenting a statusbar in the shell/tmux.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/sarifsystems/sarif/core"
	"github.com/sarifsystems/sarif/sarif"
)

var isTmux = flag.Bool("tmux", false, "tmux status line output")
var hideLabels = flag.Bool("h", false, "hide labels")
var delimiter = flag.String("d", " ", "delimiter")

func main() {
	flag.Parse()
	color.NoColor = false

	if flag.Arg(0) == "relay" {
		relay()
		return
	}

	watch()
}

type Bar struct {
	Key string `json:"name,omitempty"`

	Text string `json:"text,omitempty"`

	Label string      `json:"key,omitempty"`
	Value interface{} `json:"value,omitempty"`
	Color string      `json:"color,omitempty"`
}

func (b Bar) PrettyString(mode string) string {
	lbl := b.Label
	if lbl == "default" || *hideLabels {
		lbl = ""
	} else if lbl != "" {
		lbl += " "
	}

	text := b.Text
	if text == "" {
		if b.Value == nil {
			return ""
		}

		switch v := b.Value.(type) {
		case string:
			text = v
		case bool:
			if v {
				text = colorize(mode, "green", "✓")
			} else {
				text = colorize(mode, "red", "✗")
			}
		default:
			text = fmt.Sprintf("%v", v)
		}
	}
	if text == "" {
		return ""
	}

	cl := b.Color
	if cl == "" {
		cl = "green"
	}

	return colorize(mode, "hi-white", lbl) + colorize(mode, cl, text)
}

type Cache struct {
	FilePath string `json:"-"`

	PID         int       `json:"pid,omitempty"`
	LastUpdated time.Time `json:"time,omitempty"`

	Bars map[string]Bar `json:"bars,omitempty"`
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
	return json.NewEncoder(f).Encode(c)
}

func relay() {
	app := core.NewApp("sarif", "tars")
	app.Init()

	info := Cache{FilePath: app.Config.Dir() + "/soji.json"}
	info.Bars = make(map[string]Bar)
	info.Read()

	info.PID = os.Getpid()
	app.Must(info.Write())

	c, err := app.ClientDial(sarif.ClientInfo{
		Name: "soji/" + sarif.GenerateId(),
	})
	c.HandleConcurrent = false
	app.Must(err)

	c.Subscribe("status", "", func(msg sarif.Message) {
		key := msg.Action
		if strings.HasPrefix(key, "status/") {
			key = strings.TrimPrefix(key, "status/")
		} else {
			key = "default"
		}

		bar := Bar{Key: key, Label: key, Text: msg.Text}
		if err := msg.DecodePayload(&bar); err != nil {
			msg.DecodePayload(&bar.Text)
			msg.DecodePayload(&bar.Value)
		}

		info.Bars[bar.Key] = bar
		app.Must(info.Write())
	})

	fmt.Println(os.Getpid())
	core.WaitUntilInterrupt()

	info.PID = 0
	app.Must(info.Write())

	c.Publish(sarif.CreateMessage("soji/down", nil))
}

func watch() {
	app := core.NewApp("sarif", "tars")
	app.Init()

	info := Cache{FilePath: app.Config.Dir() + "/soji.json"}
	info.Bars = make(map[string]Bar)
	info.Read()

	err := info.Read()
	if err == nil {
		if info.PID > 0 {
			var p *os.Process
			p, err = os.FindProcess(info.PID)
			if err == nil {
				err = p.Signal(syscall.Signal(0))
			}
		} else {
			err = errors.New("PID not found")
		}
	}

	if err != nil {
		fork(app)
		return
	}

	mode := ""
	if *isTmux {
		mode = "tmux"
	}

	keys := make([]string, 0, len(info.Bars))
	if len(flag.Args()) > 0 {
		keys = flag.Args()
	} else {
		for key := range info.Bars {
			keys = append(keys, key)
		}
		sort.Strings(keys)
	}

	texts := make([]string, 0)
	for _, key := range keys {
		bar, ok := info.Bars[key]
		if !ok {
			continue
		}
		text := bar.PrettyString(mode)
		if text != "" {
			texts = append(texts, text)
		}
	}
	fmt.Println(strings.Join(texts, *delimiter))
}

func fork(app *core.App) {
	fmt.Println("booting up ...")

	cmd := exec.Command(os.Args[0], "relay")
	stdout, err := cmd.StdoutPipe()
	app.Must(err)
	app.Must(cmd.Start())

	out := bufio.NewReader(stdout)
	out.ReadString('\n')
}
