// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import "testing"

var testData = map[string]string{
	"record that i did it":                        "event/new",
	"Remind me at 15:17 to do something useful":   "schedule",
	"remind me that it is happening in 5 minutes": "schedule",
	"search google for strange things":            "knowledge/query",
	"calculate sin(x)":                            "cmd/calc",

	"Record that I started to play music":               "event/new",
	"When did I last visit Big City?":                   "location/last",
	"Remind me in 6 hours to make coffee":               "schedule",
	"Remind me at 17:15":                                "schedule",
	"Create a geofence named Work at Baker Street 221b": "location/fence/create",
	"Create a geofence at Street 13a, Berlin":           "location/fence/create",
	"What is the birth day of Tuomas Holopainen?":       "knowledge/query",
}

func TestLearningParser(t *testing.T) {
	p := NewLearningParser()
	TrainDefaults(p)

	for sentence, exp := range testData {
		msg, w := p.Parse(sentence)
		if w < 2 {
			t.Log(msg, w)
			t.Log("No message found for", sentence)
		} else if msg.Action != exp {
			t.Log(msg, w)
			t.Log("expected %s, got %s", exp, msg.Action)
		}
	}
}
