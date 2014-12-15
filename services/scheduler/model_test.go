// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package scheduler

import (
	"database/sql"
	"testing"
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
)

func TestStoreRetrieve(t *testing.T) {
	// setup test database
	deps := &Dependencies{}
	core.InjectTest(deps)

	db := &sqlDatabase{deps.DB.Driver(), deps.DB.DB}
	if err := db.Setup(); err != nil {
		t.Fatal(err)
	}

	// store two tasks
	err := db.StoreTask(Task{
		Time: time.Now().Add(2 * time.Minute),
		Reply: proto.Message{
			Action: "testaction",
			Text:   "hello you",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.StoreTask(Task{
		Time: time.Now().Add(1 * time.Minute),
		Reply: proto.Message{
			Action: "test/two",
			Text:   "hello me",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// retrieve next unfinished task (should be #2)
	task, err := db.GetNextTask()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(task)
	if task.Id == 0 {
		t.Error("id not set")
	}
	if !task.FinishedOn.IsZero() {
		t.Error("expected unfinished task, not", task.FinishedOn)
	}
	if task.Reply.Action != "test/two" {
		t.Error("wrong reply action:", task.Reply.Action)
	}
	if task.Reply.Text != "hello me" {
		t.Error("wrong reply payload:", task.Reply.Text)
	}

	// mark task as finished
	task.FinishedOn = time.Now()
	if err := db.StoreTask(task); err != nil {
		t.Fatal(err)
	}

	// retrieve next unfinished task (should be #1)
	task, err = db.GetNextTask()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(task)
	if task.Id == 0 {
		t.Error("id not set")
	}
	if !task.FinishedOn.IsZero() {
		t.Error("expected unfinished task, not", task.FinishedOn)
	}
	if task.Reply.Action != "testaction" {
		t.Error("wrong reply action:", task.Reply.Action)
	}
	if task.Reply.Text != "hello you" {
		t.Error("wrong reply payload:", task.Reply.Text)
	}

	// mark task as finished
	task.FinishedOn = time.Now()
	if err := db.StoreTask(task); err != nil {
		t.Fatal(err)
	}

	// no more tasks
	if task, err := db.GetNextTask(); err != sql.ErrNoRows {
		t.Fatal(err, task)
	}
}
