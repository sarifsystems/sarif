// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package commands

import (
	"strconv"
	"time"

	"github.com/xconstruct/stark/pkg/util"
	"github.com/xconstruct/stark/proto"
)

func printTime(t time.Time) string {
	return t.Local().Format(time.RFC1123) + "\n" +
		t.UTC().Format(time.RFC1123)
}

func (s *Service) handleDate(msg proto.Message) {
	text := msg.Text
	if text == "" {
		s.ReplyText(msg, printTime(time.Now()))
		return
	}

	if t := util.ParseTime(text, time.Now()); !t.IsZero() {
		s.ReplyText(msg, printTime(t))
		return
	}

	if d, err := util.ParseDuration(text); err == nil {
		s.ReplyText(msg, printTime(time.Now().Add(d)))
		return
	}
}

func (s *Service) handleUnix(msg proto.Message) {
	text := msg.Text
	if text == "" {
		s.ReplyText(msg, strconv.FormatInt(time.Now().Unix(), 10))
		return
	}

	if u, err := strconv.ParseInt(text, 10, 64); err == nil {
		s.ReplyText(msg, printTime(time.Unix(u, 0)))
		return
	}

	if t := util.ParseTime(text, time.Now()); !t.IsZero() {
		s.ReplyText(msg, strconv.FormatInt(t.Unix(), 10))
		return
	}

	if d, err := util.ParseDuration(text); err == nil {
		s.ReplyText(msg, strconv.FormatInt(time.Now().Add(d).Unix(), 10))
		return
	}
}
