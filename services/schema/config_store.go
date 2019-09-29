// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"encoding/json"

	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services/schema/store"
)

type ConfigStore struct {
	Store      *store.Store
	ConfigName string

	Cache      *json.RawMessage
	TriedFetch bool
	ConfigDir  string
}

func NewConfigStore(c sarif.Client) *ConfigStore {
	return &ConfigStore{
		Store:      store.New(c),
		ConfigName: c.DeviceId(),
	}
}

func (c *ConfigStore) Dir() string {
	return c.ConfigDir
}

func (c *ConfigStore) Exists() bool {
	var v interface{}
	err := c.fetch(&v)
	return err == nil
}

func (c *ConfigStore) fetch(v interface{}) error {
	return c.Store.Get("config/"+c.ConfigName, v)
}

func (c *ConfigStore) Get(v interface{}) (error, bool) {
	err := c.fetch(v)
	if err == store.ErrNotFound {
		err = c.Set(v)
		return err, true
	}
	return err, err == nil
}

func (c *ConfigStore) Set(v interface{}) error {
	_, err := c.Store.Put("config/"+c.ConfigName, v)
	return err
}
