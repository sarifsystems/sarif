// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package fddb implements part of the API of http://fddb.info
package fddb

import (
	"encoding/xml"
	"errors"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

const ROOT_URL = "http://fddb.info/api/v13"

type Client struct {
	Apikey string
	user   string
	pass   string

	Lang string
	http *http.Client
}

func New(key string) *Client {
	return &Client{
		Apikey: key,
		Lang:   "de",
		http:   &http.Client{},
	}
}

func (c *Client) SetLoginInfo(user, pass string) {
	c.user = user
	c.pass = pass
}

type request struct {
	Method string
	Args   url.Values
	Auth   bool
}

func (c *Client) do(r request, result interface{}) error {
	u, err := url.Parse(ROOT_URL + "/" + r.Method + ".xml")
	if err != nil {
		return err
	}
	r.Args.Set("apikey", c.Apikey)
	r.Args.Set("lang", c.Lang)
	r.Args.Set("enableutf8", "1")
	u.RawQuery = r.Args.Encode()

	req, err := http.NewRequest("get", u.String(), nil)
	if err != nil {
		return err
	}
	if r.Auth {
		if c.user == "" || c.pass == "" {
			return errors.New("Not logged in: username or password empty")
		}
		req.SetBasicAuth(c.user, c.pass)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("Unexpected status code: " + resp.Status)
	}

	// b, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(b))

	defer resp.Body.Close()
	dec := xml.NewDecoder(resp.Body)
	dec.CharsetReader = func(label string, in io.Reader) (io.Reader, error) {
		return charset.NewReaderLabel(label, in)
	}
	if err := dec.Decode(result); err != nil {
		return err
	}
	return nil
}

type SearchResult struct {
	Items []Item `xml:"items>item"`
}

type Item struct {
	Id             int     `xml:"id"`
	ThumbSrc       string  `xml:"thumbsrc"`
	ThumbSrcLarge  string  `xml:"thumbsrclarge"`
	FoodRank       float64 `xml:"foodrank"`
	ProducerId     int     `xml:"producerid"`
	GroupId        int     `xml:"groupid"`
	ProductcodeEAN string  `xml:"productcode_ean"`

	Data        ItemData    `xml:"data"`
	Servings    []Serving   `xml:"servings>serving"`
	Description Description `xml:"description"`
}

type ItemData struct {
	Amount                float64 `xml:"amount"`
	AmountMeasuringSystem string  `xml:"amount_measuring_system"`
	AggregateState        string  `xml:"aggregate_state"`
	KJ                    float64 `xml:"kj"`
	KCal                  float64 `xml:"kcal"`
	FatGram               float64 `xml:"fat_gram"`
	FatSatGram            float64 `xml:"fat_sat_gram"`
	KhGram                float64 `xml:"kh_gram"`
	SugarGram             float64 `xml:"sugar_gram"`
	ProteinGram           float64 `xml:"protein_gram"`
	DfGram                float64 `xml:"df_gram"`
}

type Serving struct {
	ServingId  int     `xml:"serving_id"`
	Name       string  `xml:"name"`
	WeightGram float64 `xml:"weight_gram"`
}

type Description struct {
	Name             string `xml:"name"`
	Option           string `xml:"option"`
	Producer         string `xml:"producer"`
	Group            string `xml:"group"`
	ImageDescription string `xml:"imagedescription"`
}

func (d Description) FullName() string {
	if d.Option != "" {
		return html.UnescapeString(d.Name + " " + d.Option)
	}
	return html.UnescapeString(d.Name)
}

type DiaryResult struct {
	DiaryElements []DiaryElement `xml:"diaryelement"`
}

type DiaryElement struct {
	Uid  int            `xml:"diary_uid"`
	Date int64          `xml:"diary_date"`
	Type int            `xml:"diary_type"`
	Item DiaryShortItem `xml:"diaryshortitem"`
}

type DiaryShortItem struct {
	Id          int           `xml:"itemid"`
	Data        DiaryItemData `xml:"data"`
	Description Description   `xml:"description"`
}

type DiaryItemData struct {
	ItemData
	ServingAmount float64 `xml:"diary_serving_amount"`
}

func (c *Client) SearchItem(query string) (SearchResult, error) {
	v := url.Values{}
	v.Set("q", query)
	//v.Set("startfrom", "0")
	var r SearchResult
	err := c.do(request{Method: "search/item", Args: v}, &r)
	return r, err
}

func (c *Client) DiaryGetDay(day time.Time) (DiaryResult, error) {
	method := strings.Replace(day.Format("diary/get_day 02 01 2006"), " ", "_", -1)
	v := url.Values{}
	v.Set("timezone", "Europe/London")
	var r DiaryResult
	err := c.do(request{Method: method, Args: v, Auth: true}, &r)
	return r, err
}
