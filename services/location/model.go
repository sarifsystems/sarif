// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package location

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
)

type Location struct {
	Time      time.Time `json:"time"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Geohash   string    `json:"geohash"`
	Accuracy  float64   `json:"accuracy"`
	Source    string    `json:"source,omitempty"`
	Address   string    `json:"address,omitempty"`

	Distance float64 `json:"distance,omitempty"`
	Speed    float64 `json:"speed,omitempty"`
}

func (loc Location) Key() string {
	return "locations/" + loc.Time.UTC().Format(time.RFC3339Nano) + "/" + loc.Geohash
}

type BoundingBox struct {
	LatMin float64 `json:"lat_min"`
	LatMax float64 `json:"lat_max"`
	LngMin float64 `json:"lng_min"`
	LngMax float64 `json:"lng_max"`
}

func (g *BoundingBox) GetBounds() []float64 {
	return []float64{g.LatMin, g.LatMax, g.LngMin, g.LngMax}
}

func (g *BoundingBox) SetBounds(b []float64) {
	g.LatMin, g.LatMax = b[0], b[1]
	g.LngMin, g.LngMax = b[2], b[3]
}

func (b *BoundingBox) Contains(loc Location) bool {
	return b.LatMin <= loc.Latitude &&
		b.LatMax >= loc.Latitude &&
		b.LngMin <= loc.Longitude &&
		b.LngMax >= loc.Longitude
}

type BoundingBoxSlice BoundingBox

func (b *BoundingBoxSlice) UnmarshalJSON(j []byte) (err error) {
	nums := []json.Number{}
	if err := json.Unmarshal(j, &nums); err != nil {
		return err
	}
	bs := make([]float64, len(nums))
	for i, n := range nums {
		if bs[i], err = n.Float64(); err != nil {
			return err
		}
	}
	b.LatMin, b.LatMax = bs[0], bs[1]
	b.LngMin, b.LngMax = bs[2], bs[3]
	return nil
}

type Geofence struct {
	BoundingBox
	Name       string `json:"name,omitempty"`
	Address    string `json:"address,omitempty"`
	GeohashMin string `json:"geohash_min,omitempty"`
	GeohashMax string `json:"geohash_max,omitempty"`
}

func (g Geofence) Key() string {
	return "location_geofences/" + g.Name
}

func (l Location) String() string {
	ts := l.Time.Local().Format(time.RFC3339)
	if l.Address != "" {
		return l.Address + " on " + ts
	}
	return fmt.Sprintf("%.4f, %.4f on %s", l.Latitude, l.Longitude, ts)
}

func haversine(theta float64) float64 {
	return .5 * (1 - math.Cos(theta))
}

func degToRad(deg float64) float64 {
	return deg * math.Pi / 180
}

const rEarth = 6372800 // m

func HaversineDistance(p1, p2 Location) float64 {
	lat1, lng1 := degToRad(p1.Latitude), degToRad(p1.Longitude)
	lat2, lng2 := degToRad(p2.Latitude), degToRad(p2.Longitude)

	return 2 * rEarth * math.Asin(math.Sqrt(haversine(lat2-lat1)+
		math.Cos(lat1)*math.Cos(lat2)*haversine(lng2-lng1)))
}
