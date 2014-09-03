// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package util

import "time"

func FuzzyTime(t time.Time) string {
	ny, nm, nd := time.Now().Date()
	ty, tm, td := t.Date()
	if ty == ny && tm == nm && td == nd {
		return t.Format("15:04:05")
	}
	return t.Format("02 Jan 2006 at 15:04")
}
