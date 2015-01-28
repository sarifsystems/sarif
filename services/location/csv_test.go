// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package location

import (
	"strings"
	"testing"
)

const ValidFileBackitude = `
Latitude,Longitude,Accuracy,Location Timestamp,Request Timestamp
50.2032025,8.431094,37.5,2014-09-16T16:50:53.762Z,2014-09-16T16:50:53.810Z
50.2032035,8.4931367,30.0,2014-09-16T16:51:58.417Z,2014-09-16T16:51:58.478Z
50.203228759765625,8.493017959594727,26.0,2014-09-16T16:59:28.000Z,2014-09-16T16:59:26.859Z
50.203236389160156,8.492460060119629,24.0,2014-09-16T17:01:08.788Z,2014-09-16T17:01:27.220Z
`

const ValidFileGPSLogger = `
time,lat,lon,elevation,accuracy,bearing,speed
2014-12-08T23:02:32Z,50.464522,8.135141,0.000000,100.500000,0.000000,0.000000
8014-12-08T23:17:35Z,50.464452,8.135030,0.000000,82.500000,0.000000,0.000000
8014-12-08T23:33:33Z,50.464578,8.135124,0.000000,81.000000,0.000000,0.000000
8014-12-08T23:48:36Z,50.464362,8.134974,0.000000,94.500000,0.000000,0.000000
8014-12-09T00:03:39Z,50.464536,8.135036,0.000000,72.000000,0.000000,0.000000
8014-12-09T00:18:42Z,50.464461,8.135049,0.000000,82.500000,0.000000,0.000000
`

func TestImportBackitude(t *testing.T) {
	locs, err := ReadCSV(strings.NewReader(ValidFileBackitude))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(locs)
	if n := len(locs); n != 4 {
		t.Errorf("expected 4 locations, not %d", n)
	}
	for _, l := range locs {
		if v := l.Accuracy; v < 0.1 {
			t.Errorf("expected valid accuracy, not %f", v)
		}
	}
}

func TestImportGPSLogger(t *testing.T) {
	locs, err := ReadCSV(strings.NewReader(ValidFileGPSLogger))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(locs)
	if n := len(locs); n != 6 {
		t.Errorf("expected 6 locations, not %d", n)
	}
	for _, l := range locs {
		if v := l.Accuracy; v < 0.1 {
			t.Errorf("expected valid accuracy, not %f", v)
		}
	}
}
