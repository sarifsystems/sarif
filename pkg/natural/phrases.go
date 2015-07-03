// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"math/rand"
	"time"
)

type Phrasebook struct {
	Rand    *rand.Rand
	Phrases map[string][]string
}

func NewPhrasebook() *Phrasebook {
	return &Phrasebook{
		rand.New(rand.NewSource(time.Now().UnixNano())),
		DefaultPhrases,
	}
}

func (b *Phrasebook) Get(context string) string {
	phrases := b.Phrases[context]
	if phrases == nil {
		return ""
	}

	return phrases[b.Rand.Intn(len(phrases))]
}

var DefaultPhrases = map[string][]string{
	"affirmative": {
		"Yes.",
		"Yeah.",
		"Okay.",
		"Yep.",
	},
	"success": {
		"Got it.",
		"Done.",
	},
	"negative": {
		"No.",
	},
	"unknown": {
		"I don't know.",
		"I can't help you there.",
	},
	"error": {
		"Something went wrong.",
		"Something seems to be missing.",
	},
	"compliment": {
		"You are awesome!",
		"Great!",
		"Awesome!",
		"Amazing!",
	},
	"thanks": {
		"Thanks!",
		"Thank you!",
	},
	"acknowledgement": {
		"You're welcome.",
		"Don't mention it.",
		"My pleasure.",
		"No problem.",
		"No worries.",
	},
	"greeting/initial": {
		"Hey.",
		"Hi.",
		"You there?",
	},
	"greeting/reply": {
		"Hey.",
		"Hi.",
		"Hello.",
		"How can I help you?",
	},
	"wake": {
		"Good morning!",
	},
	"apologetic": {
		"Sorry.",
	},
	"stalling": {
		"Hang on.",
		"Right away.",
		"Give me a moment.",
		"Give me a second.",
		"One moment.",
	},
}
