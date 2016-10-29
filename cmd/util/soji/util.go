// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

func colorize(mode, cl, text string) string {
	if text == "" || cl == "" {
		return text
	}

	if mode == "tmux" {
		return colorizeTmux(cl, text)
	}

	c := color.Reset
	switch cl {
	case "black":
		c = color.FgBlack
	case "red":
		c = color.FgRed
	case "green":
		c = color.FgGreen
	case "yellow":
		c = color.FgYellow
	case "blue":
		c = color.FgBlue
	case "magenta":
		c = color.FgMagenta
	case "cyan":
		c = color.FgCyan
	case "white":
		c = color.FgWhite
	case "hi-black":
		c = color.FgHiBlack
	case "hi-red":
		c = color.FgHiRed
	case "hi-green":
		c = color.FgHiGreen
	case "hi-yellow":
		c = color.FgHiYellow
	case "hi-blue":
		c = color.FgHiBlue
	case "hi-magenta":
		c = color.FgHiMagenta
	case "hi-cyan":
		c = color.FgHiCyan
	case "hi-white":
		c = color.FgHiWhite
	}
	if c != color.Reset {
		text = color.New(c).SprintfFunc()(text)
	}
	return text
}

func colorizeTmux(cl, text string) string {
	bold := ""
	if strings.HasPrefix(cl, "hi-") {
		bold = ",bold"
		cl = strings.TrimPrefix(cl, "hi-")
	}
	return fmt.Sprintf("#[fg=%s]%s#[default]", cl+bold, text)
}
