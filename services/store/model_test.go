// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package store

import (
	"testing"

	"github.com/xconstruct/stark/core"
)

func TestStore(t *testing.T) {
	store := &sqlStore{}
	core.InjectTest(store)
	if err := store.Setup(); err != nil {
		t.Fatal(err)
	}

	// Test non-existing key.
	_, err := store.Get("testkey123")
	if err != ErrNoResult {
		t.Error("expected no result, got ", err)
	}

	// Insert a document.
	_, err = store.Put(Document{
		Key:   "anotherkey",
		Value: []byte(`woop`),
	})
	if err != nil {
		t.Error(err)
	}

	// Insert another document.
	_, err = store.Put(Document{
		Key:   "testkey123",
		Value: []byte(`{"something":123}`),
	})
	if err != nil {
		t.Error(err)
	}

	// Retrieve first document
	doc, err := store.Get("anotherkey")
	if err != nil {
		t.Error(err)
	}
	if string(doc.Value) != "woop" {
		t.Error("wrong value:", string(doc.Value))
	}

	// Overwrite first document
	_, err = store.Put(Document{
		Key:   "anotherkey",
		Value: []byte(`meow`),
	})
	if err != nil {
		t.Error(err)
	}

	// Retrieve updated document
	doc, err = store.Get("anotherkey")
	if err != nil {
		t.Error(err)
	}
	if string(doc.Value) != "meow" {
		t.Error("wrong value:", string(doc.Value))
	}

	// Delete first document
	if err := store.Del("anotherkey"); err != nil {
		t.Error(err)
	}

	// Test non-existing key.
	_, err = store.Get("anotherkey")
	if err != ErrNoResult {
		t.Error("expected no result, got ", err)
	}
}
