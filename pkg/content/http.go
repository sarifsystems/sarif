// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package content

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/sarifsystems/sarif/pkg/schema"
)

type httpProvider struct{}

func (p httpProvider) Get(c schema.Content) (schema.Content, error) {
	resp, err := http.Get(c.Url)
	if err != nil {
		return c, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return c, errors.New("unexpected status " + resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c, err
	}
	c.Data = body
	c.Type = http.DetectContentType(c.Data)
	return c, err
}

func (p httpProvider) Put(c schema.Content) (schema.Content, error) {
	if c.Type == "" {
		c.Type = http.DetectContentType(c.Data)
	}
	out := schema.Content{
		Url:  c.Url,
		Type: c.Type,
	}

	resp, err := http.Post(c.Url, c.Type, bytes.NewReader(c.Data))
	if err != nil {
		return out, err
	}
	if resp.StatusCode != 200 {
		return out, errors.New("unexpected status " + resp.Status)
	}
	return out, nil
}

var HttpProvider = httpProvider{}

func init() {
	Register("http", HttpProvider)
	Register("https", HttpProvider)
}
