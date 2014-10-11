package natural

import "github.com/xconstruct/stark/proto"

const defaultRegularText = `
- example: I started to play music
  fields:
    action: event/new
    subject: [I]
    verb: [play]
    object: [music]
    status: started

- example: I finished to play music
  fields:
    action: event/new
    subject: [I]
    verb: [play]
    object: [music]
    status: ended

- example: I drink coffee
  fields:
    action: event/new
    subject: I
    verb: [drink]
    object: [coffee]
    status: singular

- example: When did I last visit Big City
  fields:
    action: location/last
    address: [Big City]

- example: When did I last drink coffee
  fields:
    action: event/last
    subject: [I]
    verb: [drink]
    object: [coffee]

- example: Push this long text to phone
  fields:
    action: push/text
    text: [this long text]
    device: [phone]

- example: Remind me in 10 minutes to make coffee
  fields:
    action: schedule/duration
    duration: [10 minutes]
    text: [make coffee]

- example: Remind me to make coffee in 10 minutes
  fields:
    action: schedule/duration
    duration: [10 minutes]
    text: [make coffee]

- example: Remind me in 10 minutes that something is happening
  fields:
    action: schedule/duration
    duration: [10 minutes]
    text: [something is happening]

- example: Remind me in 10 minutes
  fields:
    action: schedule/duration
    duration: [10 minutes]

- example: What is 3 + 5
  fields:
    action: knowledge/query
    text: [3 + 5]

- example: Create a geofence named home at 221b Baker Street
  fields:
    action: location/fence/create
    name: [home]
    address: [221b Baker Street]

- example: Create a geofence at friends house
  fields:
    action: location/fence/create
    address: [friends house]

- example: What is the birth day of Tuomas Holopainen?
  fields:
    action: knowledge/query
    text: [birth day of Tuomas Holopainen]
`

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
