// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package schema

import (
	"fmt"
	"time"
)

type Location struct {
	Timestamp time.Time `json:"timestamp,omitempty"`
	Latitude  float64   `json:"latitude,omitempty"`
	Longitude float64   `json:"longitude,omitempty"`
	Accuracy  float64   `json:"accuracy,omitempty"`
	Source    string    `json:"source,omitempty"`
	Address   string    `json:"address,omitempty"`
}

func (l Location) Text() string {
	ts := l.Timestamp.Format(time.RFC1123)
	if l.Address != "" {
		return l.Address + " on " + ts
	}
	return fmt.Sprintf("%.4f, %.4f on %s", l.Latitude, l.Longitude, ts)
}

type Geofence struct {
	LatMin  float64 `json:"lat_min,omitempty"`
	LatMax  float64 `json:"lat_max,omitempty"`
	LngMin  float64 `json:"lng_min,omitempty"`
	LngMax  float64 `json:"lng_max,omitempty"`
	Name    string  `json:"name,omitempty"`
	Address string  `json:"address,omitempty"`
}

type GeofenceChange struct {
	Location Location `json:"loc,omitempty"`
	Fence    Geofence `json:"fence,omitempty"`
	Status   string   `json:"status,omitempty"`
}

func (m GeofenceChange) Text() string {
	return fmt.Sprintf("%s %ss %s.", m.Location.Source, m.Status, m.Fence.Name)
}
