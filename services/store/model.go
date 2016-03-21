// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package store

import (
	"bytes"
	"errors"
	"strconv"

	"github.com/boltdb/bolt"
)

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

type BoltStore struct {
	DB *bolt.DB
}

func OpenBolt(path string) (*BoltStore, error) {
	db, err := bolt.Open(path, 0600, nil)
	return &BoltStore{db}, err
}

func (s *BoltStore) Put(doc *Document) (*Document, error) {
	err := s.DB.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(doc.Collection))
		if err != nil {
			return err
		}
		if doc.Key == "" {
			id, _ := b.NextSequence()
			doc.Key = "id/" + strconv.FormatUint(id, 36)
		}
		return b.Put([]byte(doc.Key), doc.Value)
	})
	return doc, err
}

func (s *BoltStore) Del(collection, key string) error {
	err := s.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		if b == nil {
			return nil
		}
		return b.Delete([]byte(key))
	})
	return err
}

func (s *BoltStore) Get(collection, key string) (*Document, error) {
	var doc *Document
	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		if b == nil {
			return nil
		}

		v := b.Get([]byte(key))
		if v == nil {
			return nil
		}

		doc = &Document{
			Collection: collection,
			Key:        key,
			Value:      v,
		}
		return nil
	})
	return doc, err
}

type boltCursor struct {
	Collection string
	First      bool
	Reverse    bool
	Min        []byte
	Max        []byte

	Tx     *bolt.Tx
	Cursor *bolt.Cursor
}

func (s *BoltStore) Scan(collection, min, max string, reverse bool) (Cursor, error) {
	var err error
	c := &boltCursor{
		Collection: collection,
		First:      true,
		Reverse:    reverse,
		Min:        []byte(min),
		Max:        []byte(max),
	}

	c.Tx, err = s.DB.Begin(false)
	if err != nil {
		return nil, err
	}
	b := c.Tx.Bucket([]byte(collection))
	if b != nil {
		c.Cursor = b.Cursor()
	}
	return c, nil
}

func (c *boltCursor) Next() (doc *Document) {
	if c.Cursor == nil {
		return nil
	}

	var k, v []byte
	if c.First {
		c.First = false
		if c.Reverse {
			if len(c.Max) > 0 {
				k, v = c.Cursor.Seek(c.Max)
			}
			if k == nil {
				k, v = c.Cursor.Last()
			}
		} else {
			k, v = c.Cursor.Seek(c.Min)
		}
	} else {
		if c.Reverse {
			k, v = c.Cursor.Prev()
		} else {
			k, v = c.Cursor.Next()
		}
	}
	if k == nil {
		return nil
	}
	if c.Reverse {
		if len(c.Min) > 0 && bytes.Compare(k, c.Min) < 0 {
			return nil
		}
	} else {
		if len(c.Max) > 0 && bytes.Compare(k, c.Max) > 0 {
			return nil
		}
	}

	return &Document{
		Collection: c.Collection,
		Key:        string(k),
		Value:      v,
	}
}

func (c *boltCursor) Close() error {
	err := c.Tx.Rollback()
	c.Tx = nil
	return err
}
