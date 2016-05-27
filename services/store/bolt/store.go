// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// BoltDB driver for the store service.
package bolt

import (
	"bytes"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/sarifsystems/sarif/services/store"
)

type Store struct {
	DB *bolt.DB
}

func Open(path string) (*Store, error) {
	db, err := bolt.Open(path, 0600, nil)
	return &Store{db}, err
}

type driver struct{}

func (d driver) Open(path string) (store.Store, error) {
	return Open(path)
}

func init() {
	store.Register("bolt", driver{})
}

func (s *Store) Put(doc *store.Document) (*store.Document, error) {
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

func (s *Store) Del(collection, key string) error {
	err := s.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		if b == nil {
			return nil
		}
		return b.Delete([]byte(key))
	})
	return err
}

func (s *Store) Get(collection, key string) (*store.Document, error) {
	var doc *store.Document
	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(collection))
		if b == nil {
			return nil
		}

		v := b.Get([]byte(key))
		if v == nil {
			return nil
		}

		doc = &store.Document{
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

func (s *Store) Scan(collection, min, max string, reverse bool) (store.Cursor, error) {
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

func (c *boltCursor) Next() (doc *store.Document) {
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

	return &store.Document{
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
