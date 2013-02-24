package natural

import (
	"errors"
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

	for name, vals := range typeVals {
		if name == "action" {
			continue
		}
		msg.Data[name] = vals[0]
	}

	return msg, nil
}

func MatchesWord(word, pattern Word) int {
	if pattern == word {
		return 3
	}
	if pattern == "*" && word != "" {
		return 2
	}
	if pattern == "" {
		return 1
	}
	return -1
}

func MatchesPhrase(phrase, pattern Phrase) int {
	sum, w := 0, 0
	if w = MatchesWord(phrase.Prev, pattern.Prev); w < 0 {
		return -1
	}
	sum += w

	if w = MatchesWord(phrase.Word, pattern.Word); w < 0 {
		return -1
	}
	sum += w

	if w = MatchesWord(phrase.Next, pattern.Next); w < 0 {
		return -1
	}
	sum += w

	return sum
}

func Parse(text string) (*stark.Message, error) {
	if text[0] == '{' {
		msg, _:= stark.Decode([]byte(text))
		msg = stark.NewMessageFromTemplate(msg)
		return msg, nil
	}

	text = strings.TrimSpace(text)
	words := strings.Split(text, " ")

	var meanings []Meaning
	words = append([]string{""}, append(words, "")...)
	var prev, word string
	for _, next := range words {
		if word != "" {
			m := GetMeanings(Phrase{Word(prev), Word(word), Word(next)})
			best := GetBestMeaning(m)
			if best.Type != "" {
				if best.Value == "*" {
					best.Value = word
				}
				meanings = append(meanings, best)
			}
		}
		prev, word = word, next
	}

	msg, err := CombineMeanings(meanings)
	if err != nil {
		return nil, err
	}
	msg.Message = text

	return msg, nil
}
