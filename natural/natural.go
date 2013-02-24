package natural

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"github.com/xconstruct/stark"
)

type Word string

type Phrase struct {
	Prev Word
	Word Word
	Next Word
}

type Meaning struct {
	Phrase Phrase
	Type string
	Value string
	Weight int
}

type WeightedMeaningSlice []Meaning

func (s WeightedMeaningSlice) Len() int {
	return len(s)
}

func (s WeightedMeaningSlice) Less(i, j int) bool {
	return s[i].Weight > s[j].Weight
}

func (s WeightedMeaningSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

var actions = map[string]string{
	"mpd": "music",
	"music": "music",
	"song": "music",
}

var subactions = map[string]string{
	"play": "play",
	"pause": "pause",
	"stop": "stop",
	"next": "next",
	"prev": "prev",
}

func GetBestMeaning(meanings []Meaning) Meaning {
	best := Meaning{Weight: -1}
	for _, m := range meanings {
		if m.Weight > best.Weight {
			best = m
		}
	}
	return best
}

func isNotWildcard(r rune) bool {
	return r != '*' && r != '.'
}

func SplitActionWildcard(action string) (pure string, prev, next int) {
	first := strings.IndexFunc(action, isNotWildcard)
	last := strings.LastIndexFunc(action, isNotWildcard)
	prev, next = first - 1, len(action) - last - 2
	if prev < 0 {
		prev = 0
	}
	if next < 0 {
		next = 0
	}
	return action[first:last+1], prev, next
}

func CombineActions(actions []string) string {
	if len(actions) == 0 {
		return ""
	}

	action, prev, next := SplitActionWildcard(actions[0])
	num := prev + 1 + next
	if num == 1 {
		return action
	}

	parts := make([]string, num)
	for _, action := range actions {
		action, prev, next := SplitActionWildcard(action)
		if parts[prev] != "" {
			continue
		}
		if (prev + 1 + next) != num {
			continue
		}
		parts[prev] = action
	}

	for _, part := range parts {
		if part == "" {
			return ""
		}
	}

	return strings.Join(parts, ".")
}

func CombineMeanings(meanings []Meaning) (*stark.Message, error) {
	sort.Sort(WeightedMeaningSlice(meanings))

	typeVals := make(map[string][]string)
	for _, m := range meanings {
		if m.Type != "" {
			typeVals[m.Type] = append(typeVals[m.Type], m.Value)
		}
	}

	var action = CombineActions(typeVals["action"])
	if action == "" {
		return nil, errors.New("No action found")
	}

	msg := stark.NewMessage()
	msg.Action = action
	return msg, nil
}

func Parse(text string) *stark.Message {
	text = strings.TrimSpace(text)
	words := strings.Split(text, " ")

	var meanings []Meaning
	words = append([]string{""}, append(words, "")...)
	var prev, word string
	for _, next := range words {
		if word != "" {
			println(prev, word, next)
			m := GetMeanings(Phrase{Word(prev), Word(word), Word(next)})
			fmt.Println(m)
			best := GetBestMeaning(m)
			if best.Type != "" {
				meanings = append(meanings, best)
			}
		}
		prev, word = word, next
	}

	msg, err := CombineMeanings(meanings)
	if err != nil {
		return nil
	}
	msg.Message = text

	return msg
}
