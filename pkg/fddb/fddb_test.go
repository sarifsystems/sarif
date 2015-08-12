// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package fddb

import (
	"os"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	key := os.Getenv("FDDB_API_KEY")
	if key == "" {
		t.Skip("Skipping FDDB test client: please specify FDDB_API_KEY")
	}

	c := New(key)
	// r, err := c.SearchItem("Nutella")
	c.SetLoginInfo(os.Getenv("FDDB_USER_NAME"), os.Getenv("FDDB_PASSWORD"))
	_, err := c.DiaryGetDay(time.Now().Add(-5 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
}
