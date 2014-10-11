// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package events

import (
	"encoding/json"
	"time"

	"github.com/xconstruct/stark/util"
)

const (
	StatusSingular   = "singular"
	StatusInProgress = "in_progress"
	StatusStarted    = "started"
	StatusEnded      = "ended"
)

type Event struct {
	Id        int64                  `json:"-"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
	Subject   string                 `json:"subject"`
	Verb      string                 `json:"verb"`
	Object    string                 `json:"object"`
	Status    string                 `json:"status"`
	Source    string                 `json:"source"`
	Text      string                 `json:"-"`
	Meta      map[string]interface{} `json:"meta" sql:"-"`
	MetaRaw   []byte                 `json:"-" gorm:"column:meta"`
}

func (e *Event) BeforeSave() (err error) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	e.MetaRaw, err = json.Marshal(e.Meta)
	return
}

func (e *Event) AfterFind() (err error) {
	if len(e.MetaRaw) == 0 {
		return nil
	}
	return json.Unmarshal(e.MetaRaw, &e.Meta)
}

func (e Event) String() string {
	if e.Text == "" {
		e.Text = e.Subject + " " + e.Verb + " " + e.Object
	}
	return util.FuzzyTime(e.Timestamp) + " - " + e.Text
}
