// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package commands

import (
	"strconv"
	"time"

	"github.com/sarifsystems/sarif/pkg/util"
	"github.com/sarifsystems/sarif/sarif"
)

func printTime(t time.Time) string {
	return t.Local().Format(time.RFC1123) + "\n" +
		t.UTC().Format(time.RFC1123) + "\n" +
		t.Local().Format(time.RFC3339)
}

func (s *Service) handleDate(msg sarif.Message) {
	text := msg.Text
	if text == "" {
		s.ReplyText(msg, printTime(time.Now()))
		return
	}

	if t := util.ParseTime(text, time.Now()); !t.IsZero() {
		s.ReplyText(msg, printTime(t))
		return
	}

	d, err := util.ParseDuration(text)
	if err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	s.ReplyText(msg, printTime(time.Now().Add(d)))
}

func (s *Service) handleUnix(msg sarif.Message) {
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

	d, err := util.ParseDuration(text)
	if err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}
	s.ReplyText(msg, strconv.FormatInt(time.Now().Add(d).Unix(), 10))
}
