// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package location

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const API_URL = "https://nominatim.openstreetmap.org"

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

type GeoName struct {
	Name string `json:"name,omitempty"`
}

type GeoPlace struct {
	BoundingBox BoundingBox `json:"boundingbox,[]string"`
	Latitude    float64     `json:"lat,string"`
	Longitude   float64     `json:"lon,string"`
	Name        string      `json:"display_name"`
	Class       string      `json:"class"`
	Type        string      `json:"type"`
	Address     GeoAddress  `json:"address"`
	NameDetails GeoName     `json:"namedetails,omitempty"`
}

func (p GeoPlace) Pretty() string {
	address := ""
	if p.NameDetails.Name != "" {
		address += p.Name
	}

	if p.Address.Road != "" {
		if address != "" {
			address += ", "
		}
		address += p.Address.Road
		if p.Address.HouseNumber != "" {
			address += " " + p.Address.HouseNumber
		}
	}

	if address != "" {
		address += ", "
	}
	if p.Address.PostCode != "" {
		address += p.Address.PostCode + " "
	}
	if p.Address.Village != "" {
		address += p.Address.Village
	} else if p.Address.Town != "" {
		address += p.Address.Town
	} else if p.Address.City != "" {
		address += p.Address.City
	}

	if address == "" {
		address = p.Name
	}
	return address
}

func Geocode(query string) ([]GeoPlace, error) {
	client := &http.Client{}

	u, err := url.Parse(API_URL + "/search/" + query)
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

type reverseResponse struct {
	GeoPlace
	Error string
}

func ReverseGeocode(loc Location) (GeoPlace, error) {
	var r reverseResponse
	client := &http.Client{}

	u, err := url.Parse(API_URL + "/reverse")
	if err != nil {
		return r.GeoPlace, err
	}

	v := url.Values{}
	v.Set("format", "json")
	v.Set("addressdetails", "1")
	v.Set("namedetails", "1")
	v.Set("zoom", "18")
	v.Set("lat", strconv.FormatFloat(loc.Latitude, 'f', -1, 64))
	v.Set("lon", strconv.FormatFloat(loc.Longitude, 'f', -1, 64))
	u.RawQuery = v.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	fmt.Println(u.String())
	req.Header.Set("User-Agent", "github.com/xconstruct/stark")
	if err != nil {
		return r.GeoPlace, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return r.GeoPlace, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return r.GeoPlace, errors.New("geocode: unexpected status " + resp.Status)
	}

	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&r)
	if r.Error != "" {
		err = errors.New("geocode: " + r.Error)
	}

	return r.GeoPlace, err
}
