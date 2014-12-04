// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import "github.com/xconstruct/stark/proto"

const defaultRegularText = `[
{
	"example": "Record that [I] started to [play] [music]",
	"msg": {
		"action": "event/new",
		"p": {
			"subject": "{{.I}}",
			"verb": "{{.play}}",
			"object": "{{.music}}",
			"status": "started"
		},
		"text": "{{.I}} started to {{.play}} {{.music}}."
	}
},
{
	"example": "Record that [I] finished to [play] [music]",
	"msg": {
		"action": "event/new",
		"p": {
			"subject": "{{.I}}",
			"verb": "{{.play}}",
			"object": "{{.music}}",
			"status": "ended"
		},
		"text": "{{.I}} finished to {{.play}} {{.music}}."
	}
},
{
	"example": "Record that [I] [drink] [coffee]",
	"msg": {
		"action": "event/new",
		"p": {
			"subject": "{{.I}}",
			"verb": "{{.drink}}",
			"object": "{{.coffee}}",
			"status": "singular"
		},
		"text": "{{.I}} {{.drink}} {{.coffee}}."
	}
},
{
	"example": "Record that [I] [work]",
	"msg": {
		"action": "event/new",
		"p": {
			"subject": "{{.I}}",
			"verb": "{{.worked}}",
			"status": "singular"
		},
		"text": "{{.I}} {{.work}}."
	}
},
{
	"example": "When did I last visit [Big City]",
	"msg": {
		"action": "location/last",
		"p": {
			"address": "{{.BigCity}}"
		}
	}
},
{
	"example": "When did [I] last [drink] [coffee]",
	"msg": {
		"action": "event/last",
		"p": {
			"subject": "{{.I}}",
			"verb": "{{.drink}}",
			"object": "{{.coffee}}"
		}
	}
},
{
	"example": "Push [this long text] to [phone]",
	"msg": {
		"action": "push/text",
		"p": {
			"device": "{{.phone}}"
		}
	}
},
{
	"example": "Remind me in [some duration] to [make coffee]",
	"msg": {
		"action": "schedule/duration",
		"p": {
			"duration": "{{.someduration}}"
		},
		"text": "{{.makecoffee}}"
	}
},
{
	"example": "Remind me to [make coffee] in [duration]",
	"msg": {
		"action": "schedule/duration",
		"p": {
			"duration": "{{.someduration}}"
		},
		"text": "{{.makecoffee}}"
	}
},
{
	"example": "Remind me at [some time] to [make coffee]",
	"msg": {
		"action": "schedule/time",
		"p": {
			"time": "{{.sometime}}"
		},
		"text": "{{.makecoffee}}"
	}
},
{
	"example": "Remind me to [make coffee] at [some time]",
	"msg": {
		"action": "schedule/time",
		"p": {
			"time": "{{.sometime}}"
		},
		"text": "{{.makecoffee}}"
	}
},
{
	"example": "Remind me in [some duration] that [something is happening]",
	"msg": {
		"action": "schedule/duration",
		"p": {
			"duration": "{{.someduration}}"
		},
		"text": "{{.somethingishappening}}"
	}
},
{
	"example": "Remind me in [some duration]",
	"msg": {
		"action": "schedule/duration",
		"p": {
			"duration": "{{.someduration}}"
		}
	}
},
{
	"example": "Remind me at [some time] that [something is happening]",
	"msg": {
		"action": "schedule/time",
		"p": {
			"time": "{{.sometime}}"
		},
		"text": "{{.somethingishappening}}"
	}
},
{
	"example": "Remind me at [some time]",
	"msg": {
		"action": "schedule/time",
		"p": {
			"time": "{{.sometime}}"
		}
	}
},
{
	"example": "Create a geofence named [home] at [Baker Street 221b]",
	"msg": {
		"action": "location/fence/create",
		"p": {
			"name": "{{.home}}",
			"address": "{{.BakerStreet221b}}"
		}
	}
},
{
	"example": "Create a geofence at [friends house]",
	"msg": {
		"action": "location/fence/create",
		"p": {
			"address": "{{.friendshouse}}"
		}
	}
},
{
	"example": "What is [the birth day of Tuomas Holopainen]",
	"msg": {
		"action": "knowledge/query",
		"text": "{{.thebirthdayofTuomasHolopainen}}"
	}
},
{
	"example": "[tell] me [some things]",
	"msg": {
		"action": "cmd/{{.tell}}",
		"text": "{{.somethings}}"
	}
}
]`

var defaultRegular RegularSchemata

func ParseRegular(text string) (proto.Message, bool) {
	if defaultRegular == nil {
		var err error
		defaultRegular, err = LoadRegularSchemata(defaultRegularText)
		if err != nil {
			panic(err)
		}
	}

	return defaultRegular.Parse(text)
}
