// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"encoding/json"
	"os"
	"path"
)

type ErrConfNotFound struct {
	Section string
}

func (e ErrConfNotFound) Error() string {
	return "conf: section '" + e.Section + "' not found"
}

type Config struct {
	modified bool
	sections map[string]*json.RawMessage
}

func NewConfig() *Config {
	return &Config{
		false,
		make(map[string]*json.RawMessage),
	}
}

func (cfg *Config) Get(section string, v interface{}) error {
	raw, ok := cfg.sections[section]
	if !ok {
		return cfg.Set(section, v)
	}
	return json.Unmarshal(*raw, v)
}

func (cfg *Config) Exists(section string) bool {
	_, ok := cfg.sections[section]
	return ok
}

func (cfg *Config) Set(section string, v interface{}) error {
	raw, err := json.MarshalIndent(&v, "", "\t")
	if err != nil {
		return err
	}
	rawjson := json.RawMessage(raw)
	cfg.sections[section] = &rawjson
	cfg.modified = true
	return nil
}

func (cfg *Config) IsModified() bool {
	return cfg.modified
}

func ReadConfig(file string) (*Config, error) {
	cfg := NewConfig()
	f, err := os.Open(file)
	if err != nil {
		return cfg, err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	err = dec.Decode(&cfg.sections)
	return cfg, err
}

func WriteConfig(file string, cfg *Config) error {
	if err := os.MkdirAll(path.Dir(file), 0700); err != nil {
		return err
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	encoded, err := json.MarshalIndent(&cfg.sections, "", "\t")
	if err != nil {
		return err
	}
	_, err = f.Write(encoded)
	return err
}
