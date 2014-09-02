// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package location

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

const API_URL = "https://nominatim.openstreetmap.org/search/"

type BoundingBox []float64

func (b *BoundingBox) UnmarshalJSON(j []byte) (err error) {
	nums := []json.Number{}
	if err := json.Unmarshal(j, &nums); err != nil {
		return err
	}
	*b = make([]float64, len(nums))
	for i, n := range nums {
		if (*b)[i], err = n.Float64(); err != nil {
			return err
		}
	}
	return nil
}

type GeoAddress struct {
	HouseNumber  string `json:"house_number"`
	Building     string `json:"building"`
	Road         string `json:"road"`
	Residential  string `json:"residential"`
	Suburb       string `json:"suburb"`
	Village      string `json:"village"`
	Town         string `json:"town"`
	CityDistrict string `json:"city_district"`
	City         string `json:"city"`
	County       string `json:"county"`
	PostCode     string `json:"postcode"`
	State        string `json:"state"`
	Country      string `json:"country"`
	CountryCode  string `json:"country_code"`
	Continent    string `json:"continent"`
}

type GeoPlace struct {
	BoundingBox BoundingBox `json:"boundingbox,[]string"`
	Latitude    float64     `json:"lat,string"`
	Longitude   float64     `json:"lon,string"`
	Name        string      `json:"display_name"`
	Class       string      `json:"class"`
	Type        string      `json:"type"`
	Address     GeoAddress  `json:"address"`
}

func (p GeoPlace) Pretty() string {
	switch p.Type {
	case "house":
		return p.Address.Road + p.Address.HouseNumber + ", " + p.Address.City
	case "village":
		return p.Address.Village + ", " + strings.ToUpper(p.Address.CountryCode)
	case "administrative":
		fallthrough
	case "city":
		return p.Address.City + ", " + strings.ToUpper(p.Address.CountryCode)
	default:
		return p.Name
	}
}

func Geocode(query string) ([]GeoPlace, error) {
	client := &http.Client{}

	u, err := url.Parse(API_URL + query)
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("format", "json")
	v.Set("addressdetails", "1")
	u.RawQuery = v.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("User-Agent", "github.com/xconstruct/stark")
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("geocode: unexpected status " + resp.Status)
	}

	results := make([]GeoPlace, 0)
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&results)
	return results, err
}
