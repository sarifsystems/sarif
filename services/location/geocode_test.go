// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package location

import "testing"

func TestGeocode(t *testing.T) {
	rs, err := Geocode("Hannover, Germany")
	if err != nil {
		t.Fatal(err)
	}

	first := rs[0]
	t.Log(first)
	if len(first.BoundingBox) != 4 {
		t.Error("bounding box has invalid size")
	}
	if first.Type != "city" {
		t.Error("expected city, not", first.Type)
	}
}
