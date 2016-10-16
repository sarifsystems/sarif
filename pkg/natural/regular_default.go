// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

var DefaultRules = SentenceRuleSet{
	"ping": "ping",
	"associate [sentence] with [action]": "natural/learn",
	"parse [text]":                       "natural/parse",

	"record that [text]": "event/new",
	"#[_action]":         "tagged",
	"#[_action] [text]":  "tagged",

	"remind me in [duration] to [text]": "schedule",
}
