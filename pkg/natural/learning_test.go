// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/xconstruct/stark/proto"
)

func TestLearningParser(t *testing.T) {
	p := NewLearningParser()

	p.LearnMessage(proto.CreateMessage("location/last", map[string]string{
		"address": "Berlin, Germany",
	}))
	p.LearnMessage(proto.CreateMessage("location/fence/create", map[string]string{
		"address": "Munich, Germany",
	}))
	p.LearnMessage(proto.CreateMessage("cmd/increment", map[string]string{
		"text": "coffees consumed",
	}))
	p.LearnMessage(proto.CreateMessage("schedule/duration", map[string]string{
		"duration": "5 hours 20 minutes",
	}))

	p.LearnSentence("remind me in [duration] to [text]")
	p.LearnSentence("increment [text]")
	p.LearnSentence("create a geofence named [name]")
	p.LearnSentence("create a geofence named [name] at [address]")

	msg, _ := p.Parse("create a geofence named test at somewhere")
	j, _ := json.Marshal(msg)
	fmt.Println(string(j))
}
