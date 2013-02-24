package natural

var dbMeanings = []Meaning{
	// Verbs
	Meaning{
		Phrase{"", "play", ""},
		"action", "*.play", 0,
	},
	Meaning{
		Phrase{"", "pause", ""},
		"action", "*.pause", 0,
	},
	Meaning{
		Phrase{"", "stop", ""},
		"action", "*.stop", 0,
	},
	Meaning{
		Phrase{"", "next", ""},
		"action", "*.next", 0,
	},
	Meaning{
		Phrase{"", "prev", ""},
		"action", "*.prev", 0,
	},
	Meaning{
		Phrase{"", "remind", ""},
		"action", "remind.*", 0,
	},

	// Nouns
	Meaning{
		Phrase{"", "music", ""},
		"action", "music.*", 0,
	},
	Meaning{
		Phrase{"", "song", ""},
		"action", "music.*", 0,
	},
	Meaning{
		Phrase{"", "mpd", ""},
		"action", "music.*", 0,
	},

	// Wildcard objects
	Meaning{
		Phrase{"", "in", "*"},
		"action", "*.in", 0,
	},
	Meaning{
		Phrase{"in", "*", ""},
		"duration", "*", 0,
	},
}

func GetMeanings(phrase Phrase) []Meaning {
	meanings := make([]Meaning, 0)
	for _, meaning := range dbMeanings {
		if MatchesPhrase(phrase, meaning.Phrase) {
			meanings = append(meanings, meaning)
			continue
		}
	}
	return meanings
}
