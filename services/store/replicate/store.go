// Copyright (C) 2019 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Replicate driver for writing to two drivers
package replicate

import (
	"errors"
	"strings"

	"github.com/sarifsystems/sarif/services/store"
)

type Store struct {
	store.Store
	Replica store.Store
}

func Open(main, replica store.Store) (*Store, error) {
	return &Store{main, replica}, nil
}

type driver struct{}

func (d driver) Open(path string) (store.Store, error) {
	dsn := strings.SplitN(path, ",", 2)

	main, err := openStoreDSN(dsn[0])
	if err != nil {
		return nil, err
	}

	replica, err := openStoreDSN(dsn[1])
	if err != nil {
		return nil, err
	}

	return Open(main, replica)
}

func openStoreDSN(dsn string) (store.Store, error) {
	parts := strings.SplitN(dsn, "=", 2)
	if len(parts) != 2 {
		return nil, errors.New("Malformed DSN")
	}

	drv, ok := store.GetDriver(parts[0])
	if !ok {
		return nil, errors.New("Unknown store driver: " + parts[0])
	}

	return drv.Open(parts[1])
}

func init() {
	store.Register("replicate", driver{})
}

func (s *Store) Put(doc *store.Document) (*store.Document, error) {
	doc, err := s.Store.Put(doc)
	if err != nil {
		return doc, err
	}

	return s.Replica.Put(doc)
}
