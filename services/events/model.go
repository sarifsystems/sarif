// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package events

import (
	"encoding/json"
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

	Timestamp time.Time              `json:"timestamp,omitempty"`
	Value     float64                `json:"value"`
	Action    string                 `json:"action,omitempty"`
	Source    string                 `json:"source,omitempty"`
	Text      string                 `json:"text,omitempty"`
	Meta      map[string]interface{} `json:"meta,omitempty" sql:"-" gorm:"column:meow"`
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
		e.Text = fmt.Sprintf("%s is %g", e.Action, e.Value)
	}
	return e.Timestamp.Format(time.RFC3339) + " - " + e.Text
}
