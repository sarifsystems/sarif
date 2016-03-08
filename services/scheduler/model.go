// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package scheduler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/xconstruct/stark/proto"
)

type Task struct {
	Id        int64         `json:"-"`
	Time      time.Time     `json:"time,omitempty"`
	Location  string        `json:"location,omitempty"`
	Reply     proto.Message `json:"reply,omitempty" sql:"-" gorm:"column:meow"`
	ReplyRaw  []byte        `json:"-" gorm:"column:reply"`
	Finished  bool          `json:"finished"`
	CreatedAt time.Time     `json:"-"`
	UpdatedAt time.Time     `json:"-"`
}

func (t Task) TableName() string {
	return "scheduler_tasks"
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

func (t *Task) BeforeSave(scope *gorm.Scope) (err error) {
	t.ReplyRaw, err = json.Marshal(t.Reply)
	scope.SetColumn("ReplyRaw", t.ReplyRaw)
	return
}

func (t *Task) AfterFind() (err error) {
	if len(t.ReplyRaw) == 0 {
		return nil
	}
	return json.Unmarshal(t.ReplyRaw, &t.Reply)
}
