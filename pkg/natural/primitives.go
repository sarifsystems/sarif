// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"fmt"
	"sort"

	"github.com/xconstruct/stark/proto"
)

type Context struct {
	Text          string
	ExpectedReply string
	Sender        string
	Recipient     string
}

type ParseResult struct {
	Text      string      `json:"text"`
	Intents   []*Intent   `json:"intents"`
	ExtraInfo interface{} `json:"extra_info"`
}

func (r ParseResult) String() string {
	s := "Interpretation: " + r.Text
	if len(r.Intents) > 0 {
		s += "\n"
		for _, intent := range r.Intents {
			s += "\n" + intent.String()
		}
	}
	return s
}

type intentByWeight []*Intent

func (a intentByWeight) Len() int           { return len(a) }
func (a intentByWeight) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a intentByWeight) Less(i, j int) bool { return a[i].Weight > a[j].Weight }

func (r *ParseResult) Merge(other *ParseResult, weight float64) {
	if other.Intents != nil {
		for _, intent := range other.Intents {
			intent.Weight *= weight

			r.Intents = append(r.Intents, intent)
		}
		sort.Sort(intentByWeight(r.Intents))
	}
}

type Intent struct {
	Intent string  `json:"intent"`
	Type   string  `json:"type"`
	Weight float64 `json:"weight"`

	Message   proto.Message `json:"msg"`
	ExtraInfo interface{}   `json:"extra_info"`
}

func (p Intent) String() string {
	s := "Intent: " + p.Intent

	v := make(map[string]string)
	p.Message.DecodePayload(&v)
	for name, val := range v {
		s += " " + name + "=" + val
	}

	s += "\n       "
	s += fmt.Sprintf(" [type: %s] [weight: %g]", p.Type, p.Weight)
	return s
}
