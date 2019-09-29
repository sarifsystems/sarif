// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package store

import (
	"errors"

	"github.com/sarifsystems/sarif/sarif"
)

var (
	ErrNotFound = errors.New("Document not found.")
	ErrNoReply  = errors.New("no reply received.")
)

type Document struct {
	Collection string `json:"collection"`
	Key        string `json:"key"`
	Value      []byte `json:"value,omitempty"`
}

type Store struct {
	client    sarif.Client
	StoreName string
}

func New(c sarif.Client) *Store {
	return &Store{client: c}
}

func checkErr(reply sarif.Message, ok bool) error {
	if !ok {
		return ErrNoReply
	}
	if reply.IsAction("err/notfound") {
		return ErrNotFound
	}
	if reply.IsAction("err") {
		return errors.New(reply.Text)
	}
	return nil
}

func (s *Store) Put(key string, doc interface{}) (*Document, error) {
	res := &Document{}
	req := sarif.CreateMessage("store/put/"+key, doc)
	req.Destination = s.StoreName
	reply, ok := <-s.client.Request(req)
	if err := checkErr(reply, ok); err != nil {
		return nil, err
	}

	return res, reply.DecodePayload(res)
}

func (s *Store) Get(key string, result interface{}) error {
	req := sarif.CreateMessage("store/get/"+key, nil)
	req.Destination = s.StoreName
	reply, ok := <-s.client.Request(req)
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
	req := sarif.CreateMessage("store/scan/"+key, &p)
	req.Destination = s.StoreName
	reply, ok := <-s.client.Request(req)
	if err := checkErr(reply, ok); err != nil {
		return err
	}

	return reply.DecodePayload(result)
}

type Command struct {
	Type  string      `json:"type"`
	Key   string      `json:"key"`
	Value interface{} `json:"value,omitempty"`
}

func (s *Store) Batch(p []Command, result interface{}) error {
	req := sarif.CreateMessage("store/batch", &p)
	req.Destination = s.StoreName
	reply, ok := <-s.client.Request(req)
	if err := checkErr(reply, ok); err != nil {
		return err
	}

	return reply.DecodePayload(result)
}

type Triple interface {
	Triple() (s, p, o string)
}

func (st *Store) PutTriple(collection string, t Triple) error {
	s, p, o := t.Triple()
	cmds := []Command{
		{"put", collection + "/spo::" + s + "::" + p + "::" + o, t},
		{"put", collection + "/sop::" + s + "::" + o + "::" + p, t},
		{"put", collection + "/ops::" + o + "::" + p + "::" + s, t},
		{"put", collection + "/osp::" + o + "::" + s + "::" + p, t},
		{"put", collection + "/pso::" + p + "::" + s + "::" + o, t},
		{"put", collection + "/pos::" + p + "::" + o + "::" + s, t},
	}
	req := sarif.CreateMessage("store/batch", &cmds)
	req.Destination = st.StoreName
	reply, ok := <-st.client.Request(req)
	if err := checkErr(reply, ok); err != nil {
		return err
	}
	return nil
}

func (st *Store) ScanTriple(collection string, t Triple, scan Scan, result interface{}) error {
	s, p, o := t.Triple()
	pre, suf := "", ""
	if s != "" {
		pre += "s"
		suf += "::" + s
	} else {
		suf = "s" + suf
	}
	if p != "" {
		pre += "p"
		suf += "::" + p
	} else {
		suf = "p" + suf
	}
	if o != "" {
		pre += "o"
		suf += "::" + o
	} else {
		suf = "o" + suf
	}

	return st.Scan(collection+"/"+pre+suf, scan, result)
}
