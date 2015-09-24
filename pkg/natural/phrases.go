// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"math/rand"
	"strings"
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

func (b *Phrasebook) GetReverse(phrase string) string {
	contexts := make(map[string]float64)

	words := SplitWords(strings.ToLower(phrase))
	for ctx, phrases := range b.Phrases {
		for _, p := range phrases {
			if count := countMatches(words, SplitWords(strings.ToLower(p))); count > 0 {
				contexts[ctx] += 1 / float64(count)
			}
		}
	}

	bestCtx, max := "", float64(0)
	for ctx, v := range contexts {
		if v > max {
			bestCtx = ctx
		}
	}
	return bestCtx
}

func (b *Phrasebook) Answer(phrase string) string {
	ctx := b.GetReverse(phrase)
	ctx2 := DefaultPhraseResponses[ctx]
	if ctx2 == "" {
		ctx2 = "unknown"
	}
	return b.Get(ctx2)
}

func countMatches(a, b []string) int {
	c := 0
	for _, w := range a {
		for _, w2 := range b {
			if w == w2 {
				c++
			}
		}
	}
	return c
}

var DefaultPhrases = map[string][]string{
	"affirmative": {
		"Yes.",
		"Yeah.",
		"Okay.",
		"Yep.",
		"Sure.",
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
		"Nice!",
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
		"How's it going?",
		"What's up?",
	},
	"greeting/reply": {
		"Hey.",
		"Hi.",
		"Hello.",
		"How can I help you?",
		"What's up?",
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
	"farewell": {
		"Bye!",
		"Have a good day!",
		"Catch you later!",
		"See you!",
	},
}

var DefaultPhraseResponses = map[string]string{
	"compliment":       "thanks",
	"thanks":           "acknowledgement",
	"greeting/initial": "greeting/reply",
	"greeting/reply":   "greeting/reply",
	"wake":             "greeting/reply",
	"farewell":         "farewell",
	"success":          "compliment",
	"affirmative":      "compliment",
}
