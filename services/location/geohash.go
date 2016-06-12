// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
// Based on Tomi Hiltunen's work (https://github.com/TomiHiltunen/geohash-golang)
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package location

import "bytes"

var (
	bits   = []int{16, 8, 4, 2, 1}
	base32 = []byte("0123456789bcdefghjkmnpqrstuvwxyz")
)

func DecodeGeohash(geohash string) *BoundingBox {
	isEven := true
	lat := []float64{-90, 90}
	lng := []float64{-180, 180}
	latErr := float64(90)
	lngErr := float64(180)

	var c string
	var cd int
	for i := 0; i < len(geohash); i++ {
		c = geohash[i : i+1]
		cd = bytes.Index(base32, []byte(c))
		for j := 0; j < 5; j++ {
			if isEven {
				lngErr /= 2
				lng = refineInterval(lng, cd, bits[j])
			} else {
				latErr /= 2
				lat = refineInterval(lat, cd, bits[j])
			}
			isEven = !isEven
		}
	}

	return &BoundingBox{lat[0], lat[1], lng[0], lng[1]}
}

func EncodeGeohash(latitude, longitude float64, precision int) string {
	isEven := true
	lat := []float64{-90, 90}
	lng := []float64{-180, 180}
	bit := 0
	ch := 0
	var geohash bytes.Buffer
	var mid float64
	for geohash.Len() < precision {
		if isEven {
			mid = (lng[0] + lng[1]) / 2
			if longitude > mid {
				ch |= bits[bit]
				lng[0] = mid
			} else {
				lng[1] = mid
			}
		} else {
			mid = (lat[0] + lat[1]) / 2
			if latitude > mid {
				ch |= bits[bit]
				lat[0] = mid
			} else {
				lat[1] = mid
			}
		}
		isEven = !isEven
		if bit < 4 {
			bit++
		} else {
			geohash.WriteByte(base32[ch])
			bit = 0
			ch = 0
		}
	}
	return geohash.String()
}

func refineInterval(interval []float64, cd, mask int) []float64 {
	if cd&mask > 0 {
		interval[0] = (interval[0] + interval[1]) / 2
	} else {
		interval[1] = (interval[0] + interval[1]) / 2
	}
	return interval
}
