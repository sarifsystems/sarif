// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package store

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestBoltStore(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "bolt")
	if err != nil {
		t.Fatal(err)
	}
	fname := f.Name()
	f.Close()
	defer os.Remove(fname)

	store, err := OpenBolt(fname)
	if err != nil {
		t.Fatal(err)
	}

	// Test non-existing key.
	doc, err := store.Get("default", "testkey123")
	if err != nil {
		t.Fatal(err)
		t.Error("expected no result, got ", err)
	}
	if doc != nil {
		t.Errorf("expected no document, got %+v", doc)
	}

	// Insert a document.
	_, err = store.Put(&Document{
		Collection: "default",
		Key:        "anotherkey",
		Value:      []byte(`woop`),
	})
	if err != nil {
		t.Error(err)
	}

	// Insert another document.
	_, err = store.Put(&Document{
		Collection: "default",
		Key:        "testkey123",
		Value:      []byte(`{"something":123}`),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve first document
	doc, err = store.Get("default", "anotherkey")
	if err != nil {
		t.Fatal(err)
	}
	if string(doc.Value) != "woop" {
		t.Error("wrong value:", string(doc.Value))
	}

	// Overwrite first document
	_, err = store.Put(&Document{
		Collection: "default",
		Key:        "anotherkey",
		Value:      []byte(`meow`),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Retrieve updated document
	doc, err = store.Get("default", "anotherkey")
	if err != nil {
		t.Error(err)
	}
	if string(doc.Value) != "meow" {
		t.Error("wrong value:", string(doc.Value))
	}

	// Delete first document
	if err := store.Del("default", "anotherkey"); err != nil {
		t.Fatal(err)
	}

	// Test non-existing key.
	doc, err = store.Get("default", "anotherkey")
	if err != nil {
		t.Fatal(err)
	}
	if doc != nil {
		t.Errorf("expected no document, got %+v", doc)
	}
}
