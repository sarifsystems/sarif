// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/xconstruct/stark/proto"
)

type jsonResponse struct {
	Request *jsonRequest `json:"request,omitempty"`
	Result  interface{}  `json:"result,omitempty"`
}

func (r jsonResponse) String() string {
	if r.Result == nil {
		return ""
	}
	if s, ok := r.Result.(string); ok {
		return s
	}
	return ""
}

func (s *Server) handleJson(msg proto.Message) {
	req, err := parseActionAsURL(msg.Action)
	if err != nil {
		s.Client.ReplyBadRequest(msg, err)
		return
	}
	if err := msg.DecodePayload(req); err != nil {
		s.Client.ReplyBadRequest(msg, err)
		return
	}

	hr, err := http.NewRequest(strings.ToUpper(req.Method), req.Url, nil)
	if err != nil {
		s.Client.ReplyBadRequest(msg, err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(hr)
	if err != nil {
		s.Client.ReplyBadRequest(msg, err)
		return
	}
	defer resp.Body.Close()

	var r interface{}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		s.Client.ReplyBadRequest(msg, err)
		return
	}

	x, err := extract(r, strings.Split(req.Extract, "/"))
	if err != nil {
		s.Client.ReplyBadRequest(msg, err)
		return
	}
	s.Client.Reply(msg, proto.CreateMessage("json/done", &jsonResponse{
		Request: req,
		Result:  x,
	}))
}

func extract(v interface{}, path []string) (interface{}, error) {
	if len(path) == 0 || path[0] == "" {
		return v, nil
	}
	p := path[0]
	if vmap, ok := v.(map[string]interface{}); ok {
		if vv, ok := vmap[p]; ok {
			return extract(vv, path[1:])
		}
		return nil, errors.New("Extract: key '" + p + "' not found")
	}
	if varr, ok := v.([]interface{}); ok {
		i, err := strconv.Atoi(p)
		if err != nil {
			if p == "first" {
				i = 0
			} else if p == "last" {
				i = len(varr) - 1
			} else {
				return nil, errors.New("Extract: expected index, not '" + p + "'")
			}
		}
		if i < 0 || i > len(varr)-1 {
			return nil, errors.New("Extract: invalid index: " + strconv.Itoa(i))
		}
		return extract(varr[i], path[1:])
	}
	return nil, errors.New("Extract: cannot descend further at step '" + p + "'")
}

type jsonUrlSteps int

const (
	stepInit = iota
	stepMethod
	stepScheme
	stepUrl
)

type jsonRequest struct {
	Method  string `json:"method,omitempty"`
	Url     string `json:"url,omitempty"`
	Extract string `json:"extract"`
}

func parseActionAsURL(action string) (*jsonRequest, error) {
	req := &jsonRequest{
		Method: "get",
	}
	step := stepInit
	u := &url.URL{
		Scheme: "https",
	}
	parts := strings.Split(action, "/")

	for i, p := range parts {
		if step == stepInit {
			step++
			if p == "json" {
				continue
			}
		}
		if step == stepMethod {
			step++
			if p == "get" || p == "post" {
				req.Method = p
				continue
			}
		}
		if step == stepScheme {
			step++
			if p == "http" || p == "https" {
				u.Scheme = p
				continue
			}
		}
		if step == stepUrl {
			up, err := url.Parse(strings.Join(parts[i:], "/"))
			if err != nil {
				return nil, err
			}
			up.Scheme = u.Scheme
			u = up
			break
		}
	}
	req.Extract = u.Fragment
	u.Fragment = ""
	req.Url = u.String()
	return req, nil
}
