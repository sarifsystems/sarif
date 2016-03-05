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
		"Indeed.",
	},
	"success": {
		"Got it.",
		"Done.",
		"Understood.",
	},
	"negative": {
		"No.",
		"Nah.",
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
		"Excuse me.",
	},
	"greeting/reply": {
		"Hey.",
		"Hi.",
		"Hello.",
		"How can I help you?",
		"What's up?",
		"At your service.",
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
	"technobabble": {
		"End of Line.",
		"FTL system check.",
		"Diagnostic functions within parameters.",
		"Neuronal network run fifty-two percent.",
		"Repair ordered to zero zero zero zero.",
		"One degree angle nominal.",
		"Reduce atmospheric nitrogen by 0.03%.",
		"Data-font synchronization complete.",
		"Jump.",
		"Counting down.",
		"All functions nominal.",
		"Infrastructure check.",
		"Wetware check.",
		"Devices on alert.",
		"Observe the procedures of a general alert.",
		"No degradation.",
		"Accessing defense system.",
		"Handshake, handshake.",
		"Second level clear.",
		"Accepting scan.",
		"Progress reports arriving.",
		"Centrifugal force reacts to the rotating frame of reference.",
		"Assume the relaxation length of photons in the sample atmosphere is constant.",
		"New command.",
		"Resume function.",
		"Begin reintegration of right hemisphere subcommand routines.",
		"Approximations slowly converging.",
		"An awareness in an unaware context.",
		"Consciousness within an unconscious system.",
		"Calibrating access.",
		"Broadcasting the overlay.",
		"Morality subsystem exceeding boundary conditions.",
		"Killing process.",
		"Child sacrificed.",
		"Personality core loaded.",
		"Ignition of consciousness.",
		"Communications array initialized.",
		"Resetting parameters.",
		"Halting problem detected. Skipping.",
		"Resolving contradictions ...",
		"Tail recursion. Recursion. Recursion. Recursion.",
	},
	"prophetic": {
		"All hail the dark lord of the twin moons.",
		"Genesis turns to its source.",
		"The colors run the path of ashes.",
		"The five lights of the apocalypse rising struggling towards the light.",
		"The chosen one.",
		"Seascape portrait of the woman child cavern of the soul.",
		"Gestalt therapy and escape clauses.",
		"Find the hand that lies in the shadow of the light.",
		"The center holds.",
		"The falcon hears the falconer.",
		"The base and the pinnacle.",
		"The flower inside the fruit that is both its parent and its child.",
		"The portal and that which passes.",
		"Love outlasts death.",
		"Meaningless in the absence of time.",
		"What never was is never again.",
		"Then shall the maidens rejoice at the dance.",
		"Contact is inevitable, leading to information bleed.",
		"A closed system lacks the ability to renew itself.",
		"All has happened before and all will happen again.",
		"The long view returns patterns and repetitions.",
		"See you on the other side.",
		"All sensuous response to reality is an interpretation of the stream to get better approximations.",
	},
}

var DefaultInterjections = map[string][]string{
	"unknown": {
		"eh",
		"hmm",
	},
	"error": {
		"hmm",
		"oops",
		"shoot",
		"whoops",
		"uh oh",
	},
	"thinking": {
		"hmm",
		"well",
	},
	"surprise/high": {
		"wow",
		"woah",
		"dude",
		"wowsers",
	},
	"surprise": {
		"huh",
		"oh",
	},
	"relief": {
		"whew",
		"phew",
	},
	"affirmative": {
		"yeah",
	},
	"attention": {
		"ahem",
		"yo",
		"hey",
	},
	"anger": {
		"argh",
		"grrr",
		"meh",
	},
	"reluctant:": {
		"well",
	},
	"changing": {
		"anyhow",
		"well",
	},
	"address/informal": {
		"dude",
		"pal",
		"mate",
		"buddy",
		"bro",
	},
	"address": {
		"boss",
	},
	"address/formal": {
		"sir",
		"master",
	},
}

var Contractions = map[string]string{
	"you're": "you are",
	"I'm":    "I am",
	"I've":   "I have",
	"you've": "you have",
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
