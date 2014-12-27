// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package schema

import "time"

const (
	EventStatusSingular   = "singular"
	EventStatusInProgress = "in_progress"
	EventStatusStarted    = "started"
	EventStatusEnded      = "ended"
)

type Event struct {
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
}
