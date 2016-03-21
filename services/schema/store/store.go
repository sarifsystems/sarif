// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package store

import (
	"errors"

	"github.com/xconstruct/stark/proto"
)

type Document struct {
	Collection string `json:"collection"`
	Key        string `json:"key"`
	Value      []byte `json:"value,omitempty"`
}

type Store struct {
	client *proto.Client
}

func New(c *proto.Client) *Store {
	return &Store{c}
}

func checkErr(reply proto.Message, ok bool) error {
	if !ok {
		return errors.New("No reply received")
	}
	if reply.IsAction("err") {
		return errors.New(reply.Text)
	}
	return nil
}

func (s *Store) Put(key string, doc interface{}) (*Document, error) {
	res := &Document{}
	reply, ok := <-s.client.Request(proto.CreateMessage("store/put/"+key, doc))
	if err := checkErr(reply, ok); err != nil {
		return nil, err
	}

	return res, reply.DecodePayload(res)
}

func (s *Store) Get(key string, result interface{}) error {
	reply, ok := <-s.client.Request(proto.CreateMessage("store/get/"+key, nil))
	if err := checkErr(reply, ok); err != nil {
		return err
	}
	return reply.DecodePayload(result)
}

type Scan struct {
	Prefix  string `json:"prefix,omitempty"`
	Start   string `json:"start,omitempty"`
	End     string `json:"end,omitempty"`
	Only    string `json:"only,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Reverse bool   `json:"reverse,omitempty"`

	Filter interface{} `json:"filter,omitempty"`
}

func (s *Store) Scan(key string, p Scan, result interface{}) error {
	reply, ok := <-s.client.Request(proto.CreateMessage("store/scan/"+key, &p))
	if err := checkErr(reply, ok); err != nil {
		return err
	}

	return reply.DecodePayload(result)
}
