// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package proto implements the stark protocol, including client and broker.
package proto

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Partial struct {
	Raw []byte
}

func (p Partial) MarshalJSON() ([]byte, error) {
	if p.Raw == nil {
		return []byte("null"), nil
	}
	return p.Raw, nil
}

func (p *Partial) UnmarshalJSON(data []byte) error {
	if p == nil {
		return errors.New("Cannot unmarshal nil argument")
	}
	if p == nil {
		return nil
	}

	p.Raw = make([]byte, len(data))
	copy(p.Raw, data)
	return nil
}

func (p Partial) String() string {
	if p.Raw == nil {
		return ""
	}
	return string(p.Raw)
}

func (p *Partial) Decode(v interface{}) error {
	if p == nil || p.Raw == nil {
		return nil
	}

	if err := json.Unmarshal(p.Raw, v); err != nil {
		return fmt.Errorf("%s. Data: %s", err.Error(), string(p.Raw))
	}
	return nil
}

func (p *Partial) MustDecode(v interface{}) {
	if err := p.Decode(v); err != nil {
		panic(err)
	}
}

func (p *Partial) Encode(v interface{}) (err error) {
	if p == nil {
		return errors.New("Cannot encode into nil pointer")
	}

	p.Raw, err = json.Marshal(v)
	return err
}

func (p *Partial) Map() (m map[string]interface{}, err error) {
	err = p.Decode(&m)
	return m, err
}
