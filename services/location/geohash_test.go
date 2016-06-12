// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
// Based on Tomi Hiltunen's work (https://github.com/TomiHiltunen/geohash-golang)
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package location

import "testing"

type geohashTest struct {
	input  string
	output *BoundingBox
}

func TestDecodeGeohash(t *testing.T) {
	var tests = []geohashTest{
		{"d", &BoundingBox{0, 45, -90, -45}},
		{"dr", &BoundingBox{39.375, 45, -78.75, -67.5}},
		{"dr1", &BoundingBox{39.375, 40.78125, -77.34375, -75.9375}},
		{"dr12", &BoundingBox{39.375, 39.55078125, -76.9921875, -76.640625}},
	}

	for _, test := range tests {
		box := DecodeGeohash(test.input)
		if *test.output != *box {
			t.Errorf("expected bounding box %v, got %v", test.output, box)
		}
	}
}

type encodeTest struct {
	lat     float64
	lng     float64
	geohash string
}

func TestEncodeGeohash(t *testing.T) {
	var tests = []encodeTest{
		{39.55078125, -76.640625, "dr12zzzzzzzz"},
		{39.5507, -76.6406, "dr18bpbp88fe"},
		{39.55, -76.64, "dr18bpb7qw65"},
		{39, -76, "dqcvyedrrwut"},
	}

	for _, test := range tests {
		geohash := EncodeGeohash(test.lat, test.lng, 12)
		if test.geohash != geohash {
			t.Errorf("expectd %s, got %s", test.geohash, geohash)
		}
	}

	for prec := range []int{3, 4, 5, 6, 7, 8} {
		for _, test := range tests {
			geohash := EncodeGeohash(test.lat, test.lng, prec)
			if len(geohash) != prec {
				t.Errorf("expected len %d, got %d", prec, len(geohash))
			}
			if test.geohash[0:prec] != geohash {
				t.Errorf("expectd %s, got %s", test.geohash, geohash)
			}
		}
	}
}
