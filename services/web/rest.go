// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/sarifsystems/sarif/sarif"
)

func RequestToMessage(req *http.Request, urlPrefix string) (sarif.Message, error) {
	msg := sarif.Message{
		Id: sarif.GenerateId(),
	}

	if err := req.ParseForm(); err != nil {
		return msg, err
	}
	spew.Dump(req.Header)

	// Parse action from url if prefix is given
	if urlPrefix != "" && strings.HasPrefix(req.URL.Path, urlPrefix) {
		msg.Action = strings.TrimPrefix(req.URL.Path, urlPrefix)
	}

	if err := parseForm(req, &msg); err != nil {
		return msg, err
	}

	return msg, nil
}

func parseForm(req *http.Request, msg *sarif.Message) error {
	pl := make(map[string]interface{})
	for k, v := range req.Form {
		if k == "authtoken" {
			continue
		}
		if k == "_device" {
			msg.Destination = v[0]
		} else if k == "text" {
			msg.Text = strings.Join(v, "\n")
		} else if len(v) == 1 {
			pl[k] = parseFormValue(v[0])
		} else if k == "_device" {
			pl[k] = v
		}
	}
	if err := msg.EncodePayload(pl); err != nil {
		return err
	}

	return nil
}

func parseFormValue(s string) interface{} {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	if v, err := strconv.ParseFloat(s, 64); err == nil {
		return v
	}
	return s
}
