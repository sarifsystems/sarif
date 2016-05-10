// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package content provides an interface to exchange content via messages.
package content

import (
	"errors"
	"strings"

	"github.com/xconstruct/stark/pkg/schema"
)

type Provider interface {
	Get(schema.Content) (schema.Content, error)
	Put(schema.Content) (schema.Content, error)
}

var providers map[string]Provider

func Register(scheme string, p Provider) {
	if providers == nil {
		providers = make(map[string]Provider)
	}
	providers[scheme] = p
}

func GetProvider(scheme string) (Provider, error) {
	if providers == nil {
		return nil, errors.New("No content provider found for scheme " + scheme)
	}
	p, ok := providers[scheme]
	if !ok {
		return nil, errors.New("No content provider found for scheme " + scheme)
	}
	return p, nil

}

func Get(c schema.Content) (schema.Content, error) {
	p, err := GetProvider(scheme(c.Url))
	if err != nil {
		return c, err
	}
	return p.Get(c)
}

func Put(c schema.Content) (schema.Content, error) {
	p, err := GetProvider(scheme(c.Url))
	if err != nil {
		return c, err
	}
	return p.Put(c)
}

func scheme(u string) string {
	i := strings.Index(u, ":")
	if i < 0 {
		return ""
	}
	return u[0:i]
}
