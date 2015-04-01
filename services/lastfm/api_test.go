// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package lastfm

import (
	"os"
	"testing"
)

func TestApiRecentTracks(t *testing.T) {
	api := NewApi("")

	tracks, err := api.UserGetRecentTracks("xconstruct", 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if tracks.Page != 1 {
		t.Error("expected page 1, not", tracks.Page)
	}
	if len(tracks.Tracks) < 40 {
		t.Log(tracks)
		t.Error("expected 50 tracks, not ", len(tracks.Tracks))
	}

	first := tracks.Tracks[1]
	if first.Artist == "" {
		t.Log("first:", first)
		t.Error("first track has no artist")
	}
	if first.Album == "" {
		t.Log("first:", first)
		t.Error("first track has no album")
	}
	if first.Name == "" {
		t.Log("first:", first)
		t.Error("first track has no name")
	}
	if first.Url == "" {
		t.Log("first:", first)
		t.Error("first track has no url")
	}
	if first.Date == "" {
		t.Log("first:", first)
		t.Error("first track has no date")
	}
	if _, err := first.ParseDate(); err != nil {
		t.Log("first:", first)
		t.Error(err)
	}
}

func TestApiTags(t *testing.T) {
	key := os.Getenv("LASTFM_API_KEY")
	if key == "" {
		t.Skip("No LASTFM_API_KEY specified.")
	}
	api := NewApi(key)

	r, err := api.ArtistGetTopTags("Subway to Sally")
	if err != nil {
		t.Fatal(err)
	}

	genre, broad := FindGenre(r.TopTags.Tags)
	if v := "folk metal"; genre != v {
		t.Errorf("expected genre %s, got %s", v, genre)
	}
	if v := "metal"; broad != v {
		t.Errorf("expected broad genre %s, got %s", v, broad)
	}
}
