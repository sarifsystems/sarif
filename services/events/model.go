// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package events

import (
	"encoding/json"
	"time"

	"github.com/xconstruct/stark/pkg/util"
)

const (
	StatusSingular   = "singular"
	StatusInProgress = "in_progress"
	StatusStarted    = "started"
	StatusEnded      = "ended"
)

type Event struct {
	Id          int64                  `json:"-"`
	Timestamp   time.Time              `json:"timestamp,omitempty"`
	Subject     string                 `json:"subject,omitempty"`
	SubjectType string                 `json:"subject_type,omitempty"`
	Verb        string                 `json:"verb,omitempty"`
	Object      string                 `json:"object,omitempty"`
	ObjectType  string                 `json:"object_type,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Source      string                 `json:"source,omitempty"`
	Text        string                 `json:"-"`
	Meta        map[string]interface{} `json:"meta,omitempty" sql:"-"`
	MetaRaw     []byte                 `json:"-" gorm:"column:meta"`
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
