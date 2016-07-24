// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package content

import (
	"encoding/base64"
	"errors"
	"net/url"
	"strings"

	"github.com/sarifsystems/sarif/pkg/schema"
)

var ErrInvalidDataURI = errors.New("Invalid data URI")

type dataProvider struct{}

func (p dataProvider) Get(c schema.Content) (schema.Content, error) {
	schemeRest := strings.SplitN(c.Url, ":", 2)
	if len(schemeRest) < 2 || schemeRest[0] != "data" {
		return c, ErrInvalidDataURI
	}

	paramData := strings.SplitN(schemeRest[1], ",", 2)
	b64 := false
	if len(paramData) > 1 {
		params := strings.Split(paramData[0], ";")
		c.Type = params[0]
		if params[len(params)-1] == "base64" {
			b64 = true
		}
	}
	data, err := url.QueryUnescape(paramData[len(paramData)-1])
	if err != nil {
		return c, err
	}
	if b64 {
		if c.Data, err = base64.StdEncoding.DecodeString(data); err != nil {
			return c, err
		}
	} else {
		c.Data = []byte(data)
	}

	if c.Type == "" {
		c.Type = "text/plain"
	}

	return c, nil
}

func (p dataProvider) Put(c schema.Content) (schema.Content, error) {
	if c.Type == "" {
		c.Type = DetectContentType(c.Data)
	}
	out := schema.Content{
		Url:  "data:" + c.Type,
		Type: c.Type,
	}
	if strings.HasPrefix(c.Type, "text/") {
		out.Url += "," + url.QueryEscape(string(c.Data))
	} else {
		out.Url += ";base64." + url.QueryEscape(base64.StdEncoding.EncodeToString(c.Data))
	}
	return out, nil
}

var DataProvider = dataProvider{}

func PutData(data []byte) schema.Content {
	c, err := DataProvider.Put(schema.Content{
		Data: data,
	})
	if err != nil {
		panic(err)
	}
	return c
}

func init() {
	Register("data", DataProvider)
}
