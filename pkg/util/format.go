// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package util

import (
	"errors"
	"time"
)

func FuzzyTime(t time.Time) string {
	ny, nm, nd := time.Now().Date()
	ty, tm, td := t.Date()
	if ty == ny && tm == nm && td == nd {
		return t.Format("15:04:05")
	}
	return t.Format("02 Jan 2006 at 15:04")
}

var timeFormats = []string{
	time.RFC3339,
	time.RFC1123,

	"2006-01-02T15:04:05Z07:00",
	"2006-01-02 15:04",
	"2006-01-02",

	"02.01.2006 on 15:04:05",
	"02.01.2006 on 15:04",
	"02.01.2006 15:04:05",
	"02.01.2006 15:04",
	"02.01.06 on 15:04:05",
	"02.01.06 on 15:04",
	"02.01.06 15:04:05",
	"02.01.06 15:04",
	"02.01.2006",
	"02.01.06",
	"02.01.",

	"2006/01/02 on 3:04:05 PM",
	"2006/01/02 on 3:04 PM",
	"2006/01/02 3:04:05 PM",
	"2006/01/02 3:04 PM",
	"01/02 on 3:04:05 PM",
	"01/02 on 3:04 PM",
	"01/02 3:04:05 PM",
	"01/02 3:04 PM",

	"15:04:05",
	"15:04",
	"3:04:05 PM",
	"3:04 PM",
	"3 PM",
}

func ParseTime(str string, rel time.Time) time.Time {
	var t time.Time
	for _, f := range timeFormats {
		var err error
		t, err = time.ParseInLocation(f, str, rel.Location())
		if err == nil {
			break
		}
	}

	if !t.IsZero() {
		if t.Year() == 0 {
			y, mo, d := rel.Date()
			if t.YearDay() != 1 {
				_, mo, d = t.Date()
			}
			h, m, s := t.Clock()
			t = time.Date(y, mo, d, h, m, s, 0, rel.Location())
		}
		if h, m, s := t.Clock(); h == 0 && m == 0 && s == 0 {
			y, mo, d := t.Date()
			h, m, s := rel.Clock()
			t = time.Date(y, mo, d, h, m, s, 0, rel.Location())
		}
	}
	return t
}

func leadingInt(s string) (x int64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x >= (1<<63-10)/10 {
			// overflow
			return 0, "", errors.New("time: bad [0-9]*")
		}
		x = x*10 + int64(c) - '0'
	}
	return x, s[i:], nil
}

var unitMap = map[string]float64{
	"ns":      float64(time.Nanosecond),
	"us":      float64(time.Microsecond),
	"µs":      float64(time.Microsecond), // U+00B5 = micro symbol
	"μs":      float64(time.Microsecond), // U+03BC = Greek letter mu
	"ms":      float64(time.Millisecond),
	"s":       float64(time.Second),
	"sec":     float64(time.Second),
	"seconds": float64(time.Second),
	"m":       float64(time.Minute),
	"min":     float64(time.Minute),
	"minute":  float64(time.Minute),
	"minutes": float64(time.Minute),
	"h":       float64(time.Hour),
	"hour":    float64(time.Hour),
	"hours":   float64(time.Hour),
	"d":       24 * float64(time.Hour),
	"day":     24 * float64(time.Hour),
	"days":    24 * float64(time.Hour),
}

func ParseDuration(s string) (time.Duration, error) {
	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	f := float64(0)
	neg := false

	// Consume [-+]?
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}
	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, nil
	}
	if s == "" {
		return 0, errors.New("util: invalid duration " + orig)
	}
	for s != "" {
		g := float64(0) // this element of the sequence

		var x int64
		var err error

		// Consume spaces and commata.
		for s != "" && (s[0] == ' ' || s[0] == ',') {
			s = s[1:]
		}

		// The next character must be [0-9.]
		if !(s[0] == '.' || ('0' <= s[0] && s[0] <= '9')) {
			return 0, errors.New("util: invalid duration " + orig)
		}
		// Consume [0-9]*
		pl := len(s)
		x, s, err = leadingInt(s)
		if err != nil {
			return 0, errors.New("util: invalid duration " + orig)
		}
		g = float64(x)
		pre := pl != len(s) // whether we consumed anything before a period

		// Consume (\.[0-9]*)?
		post := false
		if s != "" && s[0] == '.' {
			s = s[1:]
			pl := len(s)
			x, s, err = leadingInt(s)
			if err != nil {
				return 0, errors.New("util: invalid duration " + orig)
			}
			scale := 1.0
			for n := pl - len(s); n > 0; n-- {
				scale *= 10
			}
			g += float64(x) / scale
			post = pl != len(s)
		}
		if !pre && !post {
			// no digits (e.g. ".s" or "-.s")
			return 0, errors.New("util: invalid duration " + orig)
		}

		// Consume spaces.
		for s != "" && s[0] == ' ' {
			s = s[1:]
		}

		// Consume unit.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if c == ' ' || c == ',' || c == '.' || ('0' <= c && c <= '9') {
				break
			}
		}
		if i == 0 {
			return 0, errors.New("util: missing unit in duration " + orig)
		}
		u := s[:i]
		s = s[i:]
		unit, ok := unitMap[u]
		if !ok {
			return 0, errors.New("util: unknown unit " + u + " in duration " + orig)
		}

		f += g * unit
	}

	if neg {
		f = -f
	}
	if f < float64(-1<<63) || f > float64(1<<63-1) {
		return 0, errors.New("util: overflow parsing duration")
	}
	return time.Duration(f), nil
}
