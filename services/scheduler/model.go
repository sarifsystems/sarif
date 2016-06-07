// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package scheduler

import (
	"fmt"
	"time"

	"github.com/sarifsystems/sarif/sarif"
)

type Task struct {
	Id        int64         `json:"-"`
	Time      time.Time     `json:"time,omitempty"`
	Location  string        `json:"location,omitempty"`
	Reply     sarif.Message `json:"reply,omitempty"`
	Finished  bool          `json:"finished"`
	CreatedAt time.Time     `json:"-"`
	UpdatedAt time.Time     `json:"-"`
}

func (t Task) Key() string {
	return "scheduler/task/" + t.Time.UTC().Format(time.RFC3339Nano)
}

func (t Task) String() string {
	text := t.Reply.Text
	if text == "" {
		text = t.Reply.Action
	}
	if t.Reply.Action == "schedule/finished" {
		return fmt.Sprintf("Reminder for '%s' on %s set.",
			text,
			t.Time.Local().Format(time.RFC3339),
		)
	}

	return fmt.Sprintf("Schedule task '%s' on %s.",
		text,
		t.Time.Local().Format(time.RFC3339),
	)
}
