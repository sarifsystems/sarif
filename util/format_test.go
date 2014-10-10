// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package util

import (
	"testing"
	"time"
)

func TestParseTime(t *testing.T) {
	rel := time.Date(2014, 8, 9, 13, 14, 15, 0, time.UTC)

	tests := map[string]string{
		"10.11.":         "2014-11-10 13:14:15",
		"10.11.14":       "2014-11-10 13:14:15",
		"10.11.14 17:03": "2014-11-10 17:03:00",
		"17:04":          "2014-08-09 17:04:00",
		"5 PM":           "2014-08-09 17:00:00",
	}
	for str, expStr := range tests {
		exp, err := time.Parse("2006-01-02 15:04:05", expStr)
		if err != nil {
			t.Fatal(err, expStr)
		}
		if parsed := ParseTime(str, rel); parsed != exp {
			t.Errorf("expected %s, got %s", exp, parsed)
		}
	}
}

func TestParseDuration(t *testing.T) {
	tests := map[string]time.Duration{
		"2s":                     2 * time.Second,
		"15seconds":              15 * time.Second,
		"12 days 26 hours 13min": (13*24+2)*time.Hour + 13*time.Minute,
	}
	for str, exp := range tests {
		parsed, err := ParseDuration(str)
		if err != nil {
			t.Fatal(err, parsed)
		}
		if parsed != exp {
			t.Errorf("expected %s, got %s", exp, parsed)
		}
	}
}
