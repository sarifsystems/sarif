// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package lastfm

import (
	"encoding/xml"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	UrlLastfm = "https://ws.audioscrobbler.com"
)

type Api struct {
	Url string
}

func NewApi() *Api {
	return &Api{
		UrlLastfm,
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
	u, err := url.Parse(a.Url + "/2.0/user/" + user + "/recenttracks.xml")
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
