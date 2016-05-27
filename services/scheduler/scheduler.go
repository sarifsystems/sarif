// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Service scheduler provides reminders and scheduled task messages.
package scheduler

import (
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/sarifsystems/sarif/pkg/util"
	"github.com/sarifsystems/sarif/sarif"
	"github.com/sarifsystems/sarif/services"
	"github.com/sarifsystems/sarif/services/schema/store"
)

var Module = &services.Module{
	Name:        "scheduler",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	Client *sarif.Client
}

type Scheduler struct {
	*sarif.Client
	Store *store.Store

	mutex    sync.Mutex
	timer    *time.Timer
	nextTask Task
}

func NewService(deps *Dependencies) *Scheduler {
	return &Scheduler{
		Client: deps.Client,
		Store:  store.New(deps.Client),
	}
}

func (s *Scheduler) Enable() error {
	rand.Seed(time.Now().UnixNano())
	if err := s.Subscribe("schedule", "", s.handle); err != nil {
		return err
	}
	go s.simpleCron()
	go func() {
		time.Sleep(5 * time.Second)
		s.recalculateTimer()
	}()
	return nil
}

type ScheduleMessage struct {
	RandomBefore string `json:"random_before,omitempty"`
	RandomAfter  string `json:"random_after,omitempty"`
	Time         string `json:"time,omitempty"`
	Duration     string `json:"duration,omitempty"`
	Task
}

func futureTime(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	if d := time.Now().Sub(t); d > 5*time.Minute && d < 24*time.Hour {
		return t.Add(24 * time.Hour)
	}
	return t
}

func (s *Scheduler) handle(msg sarif.Message) {
	var t ScheduleMessage
	if err := msg.DecodePayload(&t); err != nil {
		s.ReplyBadRequest(msg, err)
		return
	}

	now := time.Now()
	t.Task.Time = now
	if t.Time != "" {
		t.Task.Time = futureTime(util.ParseTime(t.Time, now))
	}
	if t.RandomAfter != "" && t.RandomBefore != "" {
		after := futureTime(util.ParseTime(t.RandomAfter, t.Task.Time))
		before := futureTime(util.ParseTime(t.RandomBefore, t.Task.Time))
		if before.Before(after) {
			after, before = before, after
		}
		maxDur := int64(before.Sub(after))
		ranDur := time.Duration(rand.Int63n(maxDur))
		t.Task.Time = after.Add(ranDur)
	}
	if t.Duration != "" {
		dur, err := util.ParseDuration(t.Duration)
		if err != nil {
			s.ReplyBadRequest(msg, err)
			return
		}
		t.Task.Time = t.Task.Time.Add(dur)
	}
	if t.Task.Reply.Action == "" {
		text := msg.Text
		if text == "" {
			text = "Reminder from " + time.Now().Format(time.RFC3339) + " finished."
		}
		t.Task.Reply = sarif.Message{
			Action:      "schedule/finished",
			Destination: msg.Source,
			Text:        text,
		}
	}
	if t.Task.Reply.CorrId == "" {
		t.Reply.CorrId = msg.Id
	}
	s.Log("info", "new task:", t)

	if _, err := s.Store.Put(t.Task.Key(), &t.Task); err != nil {
		s.ReplyInternalError(msg, err)
		return
	}

	go s.recalculateTimer()
	s.Reply(msg, sarif.CreateMessage("schedule/created", t.Task))
}

func (s *Scheduler) recalculateTimer() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}

	s.nextTask = Task{}
	var tasks []Task
	err := s.Store.Scan("scheduler/task/", store.Scan{
		Start: time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339),
		Only:  "values",
		Filter: map[string]interface{}{
			"finished": false,
		},
	}, &tasks)
	if err != nil {
		s.Log("err/internal", "recalculate: "+err.Error())
	}
	if len(tasks) == 0 {
		return
	}
	s.nextTask = tasks[0]

	dur := s.nextTask.Time.Sub(time.Now())
	s.Log("debug", "next task in: "+dur.String())
	if dur < 0 {
		s.taskFinished()
		return
	}
	s.timer = time.AfterFunc(dur, s.taskFinished)
}

func (s *Scheduler) taskFinished() {
	t := s.nextTask
	t.Finished = true
	s.Log("debug", "task finished: ", t)

	if _, err := s.Store.Put(t.Key(), &t); err != nil {
		s.Log("err/internal", "could not store finished task: "+err.Error())
	}
	s.Publish(s.nextTask.Reply)
	go s.recalculateTimer()
}

func (s *Scheduler) simpleCron() {
	for {
		now := time.Now()
		nextHour := now.Add(30 * time.Minute).Round(time.Hour)
		time.Sleep(nextHour.Sub(now))
		action := strings.ToLower(time.Now().Add(5 * time.Minute).Format("cron/15h/Mon/2d/1m"))
		s.Publish(sarif.CreateMessage(action, nil))
	}
}
