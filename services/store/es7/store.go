// Copyright (C) 2019 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Elasticsearch 7 driver for the store service.
package es7

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
	esapi "github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/sarifsystems/sarif/services/store"
	"golang.org/x/net/context"
)

type Store struct {
	Client *elasticsearch.Client
}

func Open(addresses []string) (*Store, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
	}
	es, err := elasticsearch.NewClient(cfg)
	return &Store{es}, err
}

type driver struct{}

func (d driver) Open(path string) (store.Store, error) {
	return Open([]string{path})
}

func init() {
	store.Register("es7", driver{})
}

func (s *Store) Put(doc *store.Document) (*store.Document, error) {
	req := esapi.IndexRequest{
		Index:      doc.Collection,
		DocumentID: toID(doc.Key),
		Body:       bytes.NewReader(toDocument(doc.Value)),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), s.Client)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, fmt.Errorf("Error indexing document: %s", res.Status())
	}

	return doc, nil
}

func (s *Store) Del(collection, key string) error {
	return errors.New("not supported yet by driver")
}

func (s *Store) Get(collection, key string) (*store.Document, error) {
	return nil, errors.New("not supported yet by driver")
}

func (s *Store) Scan(collection, min, max string, reverse bool) (store.Cursor, error) {
	return nil, errors.New("not supported yet by driver")
}

func toID(key string) string {
	return strings.Replace(key, "/", "__", -1)
}

func toDocument(value []byte) []byte {
	var v interface{}
	if err := json.Unmarshal(value, &v); err == nil {
		if _, ok := v.(map[string]interface{}); ok {
			return value
		}
	}

	value, _ = json.Marshal(map[string]interface{}{
		"_value": value,
	})

	return value
}
