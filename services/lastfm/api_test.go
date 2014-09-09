// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package lastfm

import "testing"

func TestApiRecentTracks(t *testing.T) {
	api := NewApi()

	tracks, err := api.UserGetRecentTracks("xconstruct", 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tracks)
	if tracks.Page != 1 {
		t.Error("expected page 1, not", tracks.Page)
	}
	if len(tracks.Tracks) != 50 {
		t.Error("expected 50 tracks, not ", len(tracks.Tracks))
	}

	first := tracks.Tracks[1]
	t.Log("first:", first)
	if first.Artist == "" {
		t.Error("first track has no artist")
	}
	if first.Album == "" {
		t.Error("first track has no album")
	}
	if first.Name == "" {
		t.Error("first track has no name")
	}
	if first.Url == "" {
		t.Error("first track has no url")
	}
	if first.Date == "" {
		t.Error("first track has no date")
	}
	d, err := first.ParseDate()
	if err != nil {
		t.Error(err)
	}
	t.Log("date:", first.Date, "parsed:", d)
}
