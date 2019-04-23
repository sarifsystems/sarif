// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package store

import "errors"

var (
	ErrNoResult = errors.New("No result found.")
)

type Document struct {
	Collection string `json:"collection"`
	Key        string `json:"key"`
	Value      []byte `json:"value,omitempty"`
}

func (doc Document) String() string {
	return "Document " + doc.Key + "."
}

type Store interface {
	Put(*Document) (*Document, error)
	Get(collection, key string) (*Document, error)
	Del(collection, key string) error
	Scan(collection, min, max string, reverse bool) (Cursor, error)
}

type Cursor interface {
	Next() *Document
	Close() error
}

type Driver interface {
	Open(name string) (Store, error)
}

var drivers = make(map[string]Driver)

func Register(name string, d Driver) {
	drivers[name] = d
}

func GetDriver(name string) (Driver, bool) {
	d, ok := drivers[name]
	return d, ok
}
