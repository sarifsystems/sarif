// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package location

import (
	"fmt"
	"math"
	"time"
)

type Location struct {
	Id        int64     `json:"-"`
	Timestamp time.Time `json:"timestamp"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Accuracy  float64   `json:"accuracy"`
	Source    string    `json:"source"`
	Address   string    `json:"address,omitempty"`
}

type Geofence struct {
	Id      int64   `json:"-"`
	LatMin  float64 `json:"lat_min"`
	LatMax  float64 `json:"lat_max"`
	LngMin  float64 `json:"lng_min"`
	LngMax  float64 `json:"lng_max"`
	Name    string  `json:"name,omitempty"`
	Address string  `json:"address,omitempty"`
}

func (g Geofence) TableName() string {
	return "location_geofences"
}

func (g *Geofence) GetBounds() []float64 {
	return []float64{g.LatMin, g.LatMax, g.LngMin, g.LngMax}
}

func (g *Geofence) SetBounds(b []float64) {
	g.LatMin, g.LatMax = b[0], b[1]
	g.LngMin, g.LngMax = b[2], b[3]
}

func (l Location) String() string {
	ts := l.Timestamp.Format(time.RFC3339)
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
