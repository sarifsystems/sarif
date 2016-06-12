// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package events

import (
	"fmt"
	"time"
)

const (
	StatusSingular   = "singular"
	StatusInProgress = "in_progress"
	StatusStarted    = "started"
	StatusEnded      = "ended"
)

type Event struct {
	Id int64 `json:"-"`

	Time   time.Time              `json:"time,omitempty"`
	Value  float64                `json:"value"`
	Action string                 `json:"action,omitempty"`
	Source string                 `json:"source,omitempty"`
	Text   string                 `json:"text,omitempty"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

func (e Event) Key() string {
	return "events/" + e.Time.UTC().Format(time.RFC3339Nano) + "/" + e.Action
}

func (e Event) String() string {
	if e.Text == "" {
		e.Text = fmt.Sprintf("%s is %g", e.Action, e.Value)
	}
	return e.Time.Local().Format(time.RFC3339) + " - " + e.Text
}
