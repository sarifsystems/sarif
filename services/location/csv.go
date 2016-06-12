// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package location

import (
	"encoding/csv"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
)

var (
	ErrWrongNumberOfCols = errors.New("Number of fields and columns differs")
	ErrNoTimestamp       = errors.New("No time found")
	ErrNoLatLng          = errors.New("No latitude or longitude found")
)

var colMappings = map[string]string{
	"latitude": "latitude",
	"lat":      "latitude",

	"longitude": "longitude",
	"long":      "longitude",
	"lon":       "longitude",
	"lng":       "longitude",

	"accuracy": "accuracy",
	"acc":      "accuracy",

	"time":               "time",
	"timestamp":          "time",
	"location timestamp": "time",
}

func parseRow(fields, row []string) (*Location, error) {
	if len(fields) != len(row) {
		return nil, ErrWrongNumberOfCols
	}

	lat, lng := false, false
	var err error
	loc := &Location{}
	for i, k := range fields {
		v := row[i]
		switch k {
		case "latitude":
			lat = true
			loc.Latitude, err = strconv.ParseFloat(v, 64)
		case "longitude":
			lng = true
			loc.Longitude, err = strconv.ParseFloat(v, 64)
		case "accuracy":
			loc.Accuracy, err = strconv.ParseFloat(v, 64)
		case "time":
			loc.Time, err = time.Parse(time.RFC3339, v)
		}
		if err != nil {
			return nil, err
		}
	}

	if !lat || !lng {
		return nil, ErrNoLatLng
	}
	if loc.Time.IsZero() {
		return nil, ErrNoTimestamp
	}
	return loc, err
}

func ReadCSV(r io.Reader) ([]*Location, error) {
	cr := csv.NewReader(r)
	cols, err := cr.Read()
	if err != nil {
		return nil, err
	}

	for i, col := range cols {
		col = strings.ToLower(col)
		cols[i] = colMappings[col]
	}

	locs := make([]*Location, 0)
	for {
		row, err := cr.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return locs, err
			}
		}

		loc, err := parseRow(cols, row)
		if err != nil {
			return nil, err
		}
		locs = append(locs, loc)
	}
	return locs, nil
}
