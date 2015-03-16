// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package lastfm

import (
	"encoding/json"
	"encoding/xml"
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
	Artist     string `xml:"artist"`
	Album      string `xml:"album"`
	Name       string `xml:"name"`
	Url        string `xml:"url"`
	Date       string `xml:"date"`
	NowPlaying bool   `xml:"nowplaying,attr"`
}

func (t ApiTrack) ParseDate() (time.Time, error) {
	return time.Parse("_2 Jan 2006, 15:04", t.Date)
}

type ApiRecentTracks struct {
	User       string     `xml:"user,attr"`
	Page       int        `xml:"page,attr"`
	PerPage    int        `xml:"perPage,attr"`
	TotalPages int        `xml:"totalPages,attr"`
	Total      int        `xml:"total,attr"`
	Tracks     []ApiTrack `xml:"track"`
}

func (a *Api) UserGetRecentTracks(user string, page int, from int64) (*ApiRecentTracks, error) {
	if page < 1 {
		page = 1
	}
	u, err := url.Parse(a.Url + "/user/" + user + "/recenttracks.xml")
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("page", strconv.Itoa(page))
	v.Set("limit", "49")
	if from > 0 {
		v.Set("from", strconv.FormatInt(from, 10))
	}
	u.RawQuery = v.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("lastfm api: unexpected status " + resp.Status)
	}

	result := &ApiRecentTracks{}
	dec := xml.NewDecoder(resp.Body)
	err = dec.Decode(result)
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
	Count int    `json:"count,string"`
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
