// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

var LearningSentencesDefault = []string{
	"record that [text]",
	"note that [text]",

	"when did i last visit [address]",
	"push [text] to [device]",
	"push [text]",

	"get last event",
	"get last event with action [action]",
	"get last event from [action]",
	"get last event where [filter]",
	"find last event",
	"find last event with action [action]",
	"find last event from [action]",
	"find last event where [filter]",

	"list events with action [action]",
	"list events where [filter]",
	"find events with action [action]",
	"find events where [filter]",

	"get last location",
	"find last location",
	"find last location with address [address]",
	"find last location at [address]",

	"remind me at [time]",
	"remind me at [time] to [text]",
	"remind me at [time] that [text]",
	"remind me in [duration]",
	"remind me in [duration] to [text]",
	"remind me in [duration] that [text]",
	"remind me that [text] in [duration]",
	"remind me that [text] at [time]",
	"remind me to [text] in [duration]",
	"remind me to [text] at [time]",
	"schedule at [time]",
	"schedule at [time] to [text]",
	"schedule at [time] that [text]",
	"schedule in [duration]",
	"schedule in [duration] to [text]",
	"schedule in [duration] that [text]",
	"schedule that [text] in [duration]",
	"schedule that [text] at [time]",
	"schedule to [text] in [duration]",
	"schedule to [text] at [time]",

	"set alarm for [time]",
	"set alarm in [duration]",
	"wake me up at [time]",
	"wake me up in [duration]",

	"create a geofence named [name]",
	"create a geofence named [name] at [address]",
	"create a geofence at [address]",

	"[counter]++",
	"[counter]--",

	"play music",
	"play [artist]",
	"play artist [artist]",
	"play music by [artist]",
	"play music by artist [artist]",
	"play album [album]",
	"play music from the album [album]",
	"play music from album [album]",
	"listen to music",
	"listen to artist [artist]",
	"listen to music by [artist]",
	"listen to music by artist [artist]",
	"listen to album [album]",
	"listen to music from the album [album]",
	"listen to music from album [album]",
	"stop music",
	"stop playing music",

	"search [provider] for [query]",
	"search for [query]",
	"what is [query]",
	"calculate [text]",
	"how much is [query]",
}

var LearningMessagesDefault = []MessageSchema{
	{
		Action: "event/new",
		Fields: map[string]string{
			"text": "string",
		},
	},
	{
		Action: "event/last",
		Fields: map[string]string{
			"filter": "string",
		},
	},
	{
		Action: "location/last",
		Fields: map[string]string{
			"address": "string",
		},
	},
	{
		Action: "location/fence/create",
		Fields: map[string]string{
			"address": "string",
			"name":    "string",
		},
	},
	{
		Action: "location/find",
		Fields: map[string]string{
			"address": "string",
		},
	},
	{
		Action: "schedule",
		Fields: map[string]string{
			"time":     "string",
			"duration": "string",
			"text":     "string",
		},
	},
	{
		Action: "knowledge/query",
		Fields: map[string]string{
			"query": "string",
			"text":  "string",
		},
	},
	{
		Action: "cmd/calc",
		Fields: map[string]string{
			"query": "string",
		},
	},
	{
		Action: "mpd/play",
		Fields: map[string]string{
			"artist": "string",
			"album":  "string",
			"title":  "string",
		},
	},
	{
		Action: "mpd/pause",
	},
	{
		Action: "mpd/stop",
	},
}

var LearningReinforcementDefaults = map[string]string{
	"record that things happen":                   "event/new",
	"remind me in 10 seconds to do something":     "schedule",
	"calculate 3 + 5":                             "cmd/calc",
	"search for 3 + 5":                            "knowledge/query",
	"get last event from mood/general":            "event/last",
	"find last event with action location/update": "event/last",
	"find events with action condition/changed":   "event/find",
	"play music by nightwish":                     "mpd/play",
	"find last location with address Berlin":      "location/last",
	"when did i last visit Berlin":                "location/last",
	"create a geofence at Berlin":                 "location/fence/create",
	"what is this thing":                          "knowledge/query",
}
