// Copyright (C) 2017 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package schema

import "time"

type MusicInfo struct {
	IsPlaying bool      `json:"is_playing"`
	Device    string    `json:"device,omitempty"`
	Time      time.Time `json:"time,omitempty"`

	Artist   string `json:"artist,omitempty"`
	Album    string `json:"album,omitempty"`
	Track    string `json:"track,omitempty"`
	Duration int    `json:"duration,omitempty"`
}
