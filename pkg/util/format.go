// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package util

import (
	"errors"
	"strings"
	"time"
)

func FuzzyTime(t time.Time) string {
	n := time.Now()
	ny, nm, nd := n.Date()
	ty, tm, td := t.Date()
	if ty == ny && tm == nm && td == nd {
		return t.Format("15:04:05")
	}
	if d := n.YearDay() - t.YearDay(); ny == ty && d > -8 && d < 8 {
		return t.Format("02 Jan 06 at 15:04")
	}

	return t.Format("Mon, 02 Jan 06 at 15:04")
}

var datetimeFormats = []string{
	time.RFC3339,
	time.RFC1123,
}

var dateFormats = []string{
	"2006-01-02",

	"02.01.2006",
	"02.01.06",
	"02.01.",
	"2.1.2006",
	"2.1.06",
	"2.1.",

	"01/02/2006",
	"01/02/06",
	"01/02",
	"1/2/2006",
	"1/2/06",
	"1/2",
}

var timeFormats = []string{
	"15:04:05",
	"15:04",
	"3:04:05 PM",
	"3:04 PM",
	"3 PM",
}

func tryAll(str string, rel time.Time, formats []string) (time.Time, bool) {
	for _, f := range formats {
		t, err := time.ParseInLocation(f, str, rel.Location())
		if err == nil {
			return t, true
		}
	}
	return rel, false
}

func ParseTime(str string, rel time.Time) time.Time {
	if t, ok := tryAll(str, rel, datetimeFormats); ok {
		return t
	}

	var dt, tt time.Time
	var dok, tok bool
	if strings.Index(str, " ") > -1 {
		parts := strings.SplitN(str, " ", 2)
		dt, dok = tryAll(parts[0], rel, dateFormats)
		tt, tok = tryAll(parts[1], rel, timeFormats)
	}
	if !dok || !tok {
		dt, dok = tryAll(str, rel, dateFormats)
		tt, tok = tryAll(str, rel, timeFormats)
	}
	if !dok && !tok {
		return time.Time{}
	}

	y, mo, d := dt.Date()
	if y == 0 {
		y, _, _ = rel.Date()
	}
	h, m, s := tt.Clock()
	return time.Date(y, mo, d, h, m, s, 0, rel.Location())
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
