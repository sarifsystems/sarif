// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package lastfm

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	UrlLastfm = "https://ws.audioscrobbler.com/2.0"
)

type Api struct {
	Url string
	Key string
}

func NewApi(key string) *Api {
	return &Api{
		UrlLastfm,
		key,
	}
}

type ApiTrack struct {
	Artist struct {
		Text string `json:"#text"`
	} `json:"artist"`
	Album struct {
		Text string `json:"#text"`
	} `json:"album"`
	Name string `json:"name"`
	Url  string `json:"url"`
	Date struct {
		UTS  int    `json:"uts,string"`
		Text string `json:"#text"`
	} `json:"date"`
	Attr struct {
		NowPlaying bool `json:"nowplaying,string"`
	} `json:"@attr"`
}

func (t ApiTrack) ParseDate() (time.Time, error) {
	return time.Parse("_2 Jan 2006, 15:04", t.Date.Text)
}

type ApiRecentTracks struct {
	ApiResponse
	RecentTracks struct {
		Attr struct {
			User       string `json:"user"`
			Page       int    `json:"page,string"`
			PerPage    int    `json:"perPage,string"`
			TotalPages int    `json:"totalPages,string"`
			Total      int    `json:"total,string"`
		} `json:"@attr"`
		Tracks []ApiTrack `json:"track"`
	} `json:"recenttracks"`
}

func (a *Api) UserGetRecentTracks(user string, page int, from int64) (*ApiRecentTracks, error) {
	result := &ApiRecentTracks{}
	v := url.Values{}
	v.Set("page", strconv.Itoa(page))
	v.Set("limit", "49")
	v.Set("user", user)
	if from > 0 {
		v.Set("from", strconv.FormatInt(from, 10))
	}
	err := a.authDo("user.getRecentTracks", v, result)
	return result, err
}

func (a *Api) authDo(method string, args url.Values, result interface{}) error {
	if a.Key == "" {
		return errors.New("no API key specified")
	}
	u, err := url.Parse(a.Url + "/")
	if err != nil {
		return err
	}
	args.Set("api_key", a.Key)
	args.Set("method", method)
	args.Set("format", "json")
	u.RawQuery = args.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("unexpected status code: " + resp.Status)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}
	if r, ok := result.(iserrorer); ok && r.IsError != nil {
		return r.IsError()
	}
	return nil
}

type iserrorer interface {
	IsError() error
}

type ApiResponse struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
}

func (r ApiResponse) IsError() error {
	if r.Error == 0 {
		return nil
	}
	return errors.New(r.Message)
}

type ApiTopTags struct {
	ApiResponse
	TopTags struct {
		Tags []ApiTag `json:"tag"`
		Attr struct {
			Artist string `json:"artist"`
		} `json:"@attr"`
	} `json:"toptags"`
}

type ApiTag struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Url   string `json:"url"`
}

func (a *Api) ArtistGetTopTags(artist string) (*ApiTopTags, error) {
	result := &ApiTopTags{}
	v := url.Values{}
	v.Set("artist", artist)
	err := a.authDo("artist.getTopTags", v, result)
	if _, ok := err.(*json.UnmarshalTypeError); ok {
		err = nil
	}
	return result, err
}

var IgnoredGenres = map[string]string{
	"composer":         "",
	"composers":        "",
	"female vocalists": "",
	"hip hop":          "hip-hop",
	"instrumental":     "",
	"rap":              "hip-hop",
}

var BroadGenres = []string{
	"acoustic",
	"classical",
	"country",
	"electronic",
	"folk",
	"gothic",
	"hip-hop",
	"indie",
	"jazz",
	"metal",
	"podcast",
	"pop",
	"punk",
	"reggae",
	"rock",
	"singer-songwriter",
	"soundtrack",
}

func FindGenre(tags []ApiTag) (string, string) {
	var genre, broad string
	for _, t := range tags {
		name := strings.ToLower(t.Name)
		if replace, ok := IgnoredGenres[name]; ok {
			if broad == "" {
				broad = replace
			}
		} else if inArray(name, BroadGenres) {
			if broad == "" {
				broad = name
			}
		} else if genre == "" {
			genre = name
		}

		if broad != "" && genre != "" {
			break
		}
	}
	return genre, broad
}

func inArray(s string, a []string) bool {
	for _, v := range a {
		if v == s {
			return true
		}
	}
	return false
}
