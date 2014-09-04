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
    subject: [I]
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

- example: Remind me in 10m to make coffee
  fields:
    action: schedule/duration
    duration: [10m]
    text: [make coffee]

- example: Remind me in 10m
  fields:
    action: schedule/duration
    duration: [10m]

- example: What is 3 + 5
  fields:
    action: know/query
    query: [3 + 5]
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
