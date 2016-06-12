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
	if first.Type != "city" {
		t.Error("expected city, not", first.Type)
	}
}

func TestReverseGeocode(t *testing.T) {
	place, err := ReverseGeocode(Location{
		Latitude:  52.5487429714954,
		Longitude: -1.81602098644987,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(place)
	if place.Address.HouseNumber != "137" {
		t.Error("wrong house number:", place.Address.HouseNumber)
	}
	if place.Address.Road != "Pilkington Avenue" {
		t.Error("wrong road:", place.Address.Road)
	}
}
