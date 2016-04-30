// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package natural contains a multitude of natural language parsers.
package natural

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/xconstruct/stark/pkg/util"
	"github.com/xconstruct/stark/proto"
)

const (
	ThisAlgorithmBecomingSkynetCost = 999999999
)

func ParseSimple(text string) (proto.Message, bool) {
	msg := proto.Message{}

	// Raw JSON message
	if strings.HasPrefix(text, "{") {
		if err := json.Unmarshal([]byte(text), &msg); err == nil {
			return msg, true
		}
	}

	if strings.HasPrefix(text, "stark://") {
		if msg, err := proto.ConvertURL(text); err == nil {
			return msg, true
		}
	}

	if strings.HasPrefix(text, ".") || strings.HasPrefix(text, "/") {
		text = strings.TrimLeft(text, "./ ")
		parts, _ := SplitQuoted(text, " ")
		if parts[0] == "" {
			return msg, false
		}
		msg.Action = parts[0]

		msg.Text = ""
		payload := make(map[string]interface{}, 0)
		for _, part := range parts[1:] {
			keyval, _ := SplitQuoted(part, "=")
			if len(keyval) == 1 {
				if msg.Text != "" {
					msg.Text += " "
				}
				msg.Text += TrimQuotes(keyval[0])
				continue
			}

			k := TrimQuotes(keyval[0])
			vtext := strings.Join(keyval[1:], "=")
			quoted := strings.ContainsAny(vtext, "\"`")
			vtext = TrimQuotes(vtext)
			var v interface{} = vtext
			if !quoted {
				v = parseValue(vtext)
			}
			switch k {
			case "text":
				msg.Text = vtext
			case "device":
				fallthrough
			case "destination":
				msg.Destination = vtext
			default:
				payload[k] = v
			}
		}
		if len(payload) > 0 {
			msg.EncodePayload(payload)
		}
		return msg, true
	}

	return msg, false
}

func parseValue(v string) interface{} {
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return f
	}
	if t := util.ParseTime(v, time.Now()); !t.IsZero() {
		return t
	}
	return v
}

func FormatSimple(msg proto.Message) string {
	if msg.Text != "" {
		return msg.Text
	}

	return fmt.Sprintf("%s from %s.", msg.Action, msg.Source)
}

var (
	reTimeIso = regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}\:\d{2}\:\d{2}[+-]\d{2}\:\d{2})`)
)

func formatTime(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}

	return util.FuzzyTime(t)
}

func FormatMessage(msg *proto.Message) {
	if msg.Text == "" {
		return
	}

	msg.Text = reTimeIso.ReplaceAllStringFunc(msg.Text, formatTime)
}

func SplitWords(s string) []string {
	s = strings.TrimRight(s, ".!? ")
	return strings.Split(s, " ")
}
