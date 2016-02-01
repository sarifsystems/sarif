// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package sparql implements a simple SPARQL client and query builder.
package sparql

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

type GenericEndpoint struct {
	Url string
}

func NewEndpoint(url string) *GenericEndpoint {
	return &GenericEndpoint{
		Url: url,
	}
}

func (p *GenericEndpoint) request(q, format string) (*http.Response, error) {
	v := url.Values{}
	v.Add("query", q)
	v.Add("format", format)

	u, err := url.Parse(p.Url)
	if err != nil {
		return nil, err
	}
	u.RawQuery = v.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return resp, err
	}

	if resp.StatusCode != http.StatusOK {
		msg, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return resp, err
		}
		return resp, errors.New("Unexpected status: " + resp.Status + "\n\n" + string(msg))
	}
	return resp, nil
}

func (p *GenericEndpoint) Select(q string, result interface{}) error {
	resp, err := p.request(q, "application/sparql-results+json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}

func (p *GenericEndpoint) Describe(q string, result interface{}) error {
	resp, err := p.request(q, "application/x-json+ld")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}

func (p *GenericEndpoint) Query() *Query {
	return New(p)
}

var (
	DBPedia = NewEndpoint("http://dbpedia.org/sparql")
)

type ResourceResponse struct {
	Results struct {
		Bindings []Row `json:"bindings"`
	} `json:"results"`
}

type Resource struct {
	Type     string `json:"type"`
	DataType string `json:"datatype"`
	Value    string `json:"value"`
}

type Row map[string]Resource
