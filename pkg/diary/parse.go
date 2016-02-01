package diary

import (
	"regexp"
	"strings"
	"unicode"

	"gopkg.in/yaml.v2"
)

type Entry struct {
	Title string `yaml:"title,omitempty"`
	Date  string `yaml:"date,omitempty"`

	TimeMood    map[string]string `yaml:"time_mood,omitempty"`
	AverageMood string            `yaml:"average_mood,omitempty"`
	Grateful    []string          `yaml:"grateful,omitempty"`

	Relevance    string   `yaml:"relevance,omitempty"`
	Achievements []string `yaml:"achievements,omitempty"`
	Event        string   `yaml:"event,omitempty"`
	Quote        string   `yaml:"quote,omitempty"`
	Nutshell     string   `yaml:"nutshell,omitempty"`
	Music        string   `yaml:"music,omitempty"`

	Version string `yaml:v,omitempty"`

	Text     string `yaml:"-"`
	FileName string `yaml:"-"`
}

func Decode(s string) (*Entry, error) {
	e := &Entry{}
	s = strings.TrimSpace(s)

	if s[0:3] == "---" {
		if err := yaml.Unmarshal([]byte(s), e); err != nil {
			return e, err
		}
		if i := strings.Index(s[3:], "---"); i >= 0 {
			s = s[6+i:]
		}
	}

	e.Text = strings.TrimSpace(s)
	return e, nil
}

func Encode(e *Entry) (string, error) {
	b, err := yaml.Marshal(e)
	if err != nil {
		return "", err
	}

	raw := "---\n" + string(b) + "\n---\n\n" + e.Text
	raw = strings.TrimSpace(raw) + "\n"
	return raw, nil
}

var reTag = regexp.MustCompile(`\s#([^#\s]+)`)

func (e Entry) Tags() map[string]int {
	counts := make(map[string]int)

	matches := reTag.FindAllStringSubmatch(" "+e.Text+" ", -1)
	for _, m := range matches {
		counts[m[1]]++
	}

	return counts
}

func (e Entry) Matches(tag string) bool {
	re := regexp.MustCompile(`(?i)\s#` + regexp.QuoteMeta(tag) + `\s`)
	if re.MatchString(" " + e.Text + " ") {
		return true
	}

	return false
}

func (e Entry) MatchesInexact(tag string) bool {
	if e.Matches(tag) {
		return true
	}

	tag = regexp.QuoteMeta(tag)
	inexact := ""
	for _, r := range tag {
		if unicode.IsUpper(r) {
			inexact += `\s?` + string(unicode.ToLower(r))
		} else {
			inexact += string(r)
		}
	}
	text := " " + e.Text + " "
	re, err := regexp.Compile(`(?i)\s` + inexact + `\W`)
	return err == nil && re.FindString(text) != ""
}
