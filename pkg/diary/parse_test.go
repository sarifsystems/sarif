package diary

import "testing"

const TestEntry = `
---
title: A Test entry
date: 2016-02-01
time_mood:
  morning:   +/10:00
  noon:      o/12:00
  afternoon: -/16:00
  evening:   o/19:00
  night:     +
grateful:
  - Thing 1
  - Thing 2
  - Thing 3
achievements:
  - Worked
  - Did something
quote: Something something testing
nutshell: I did something and it was great.
music: Sumthing
v: 1.0
---

Dear Diary,

Today OP was a good guy.
Linebreak.
#HashTag #Awesome #Things

Paragraph.
End.
#Finally #Awesome`

func TestFormat(t *testing.T) {
	e, err := Decode(TestEntry)
	if err != nil {
		t.Fatal(err)
	}

	if e.Title != "A Test entry" {
		t.Error("Title invalid!")
	}
	if len(e.Achievements) != 2 {
		t.Error("Invalid achievement list")
	}

	tags := e.Tags()
	if tags["Awesome"] != 2 {
		t.Error("tag count failed")
	}

	if !e.Matches("hashtag") {
		t.Error("Text should match hashtag")
	}
	if e.Matches("Hash") {
		t.Error("Text should not match Hash")
	}

	if !e.MatchesInexact("GoodGuy") {
		t.Error("Text should match GoodGuy")
	}
}
