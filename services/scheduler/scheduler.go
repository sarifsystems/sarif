// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package scheduler

import (
	"database/sql"
	"time"

	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
	"github.com/xconstruct/stark/util"
)

var Module = &services.Module{
	Name:        "scheduler",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	DB     *core.DB
	Log    proto.Logger
	Client *proto.Client
}

type Scheduler struct {
	timer *time.Timer
	DB    Database
	Log   proto.Logger
	*proto.Client
	nextTask Task
}

func NewService(deps *Dependencies) *Scheduler {
	return &Scheduler{
		DB:     &sqlDatabase{deps.DB.Driver(), deps.DB.DB},
		Log:    deps.Log,
		Client: deps.Client,
	}
}

func (s *Scheduler) Enable() error {
	if err := s.DB.Setup(); err != nil {
		return err
	}
	if err := s.Subscribe("schedule", "", s.handle); err != nil {
		return err
	}
	s.recalculateTimer()
	return nil
}

type ScheduleMessage struct {
	Time     string `json:"time,omitempty"`
	Duration string `json:"duration,omitempty"`
	Task
}

func (s *Scheduler) handle(msg proto.Message) {
	if !msg.IsAction("schedule") {
		return
	}

	var t ScheduleMessage
	if err := msg.DecodePayload(&t); err != nil {
		s.Log.Warnln("[scheduler] received bad payload:", err)
		s.ReplyBadRequest(msg, err)
		return
	}

	if t.Time != "" {
		t.Task.Time = util.ParseTime(t.Time, time.Now())
	}
	if t.Task.Time.IsZero() {
		t.Task.Time = time.Now()
	}
	if t.Duration != "" {
		dur, err := util.ParseDuration(t.Duration)
		if err != nil {
			s.Log.Warnln("[scheduler] received bad payload:", err)
			s.ReplyBadRequest(msg, err)
			return
		}
		t.Task.Time = t.Task.Time.Add(dur)
	}
	if t.Task.Reply.Action == "" {
		text := msg.Text
		if text == "" {
			text = "Reminder from " + util.FuzzyTime(time.Now()) + " finished."
		}
		t.Task.Reply = proto.Message{
			Action:      "schedule/finished",
			Destination: msg.Source,
			Text:        text,
		}
	}
	if t.Task.Reply.CorrId == "" {
		t.Reply.CorrId = msg.Id
	}
	s.Log.Infoln("[scheduler] new task:", t)

	if err := s.DB.StoreTask(t.Task); err != nil {
		s.Log.Errorln("[scheduler] could not store task:", err)
		s.ReplyInternalError(msg, err)
		return
	}

	reply := proto.Message{Action: "schedule/created"}
	if err := reply.EncodePayload(t); err != nil {
		s.Log.Errorln("[scheduler] could not encode reply:", err)
		s.ReplyInternalError(msg, err)
		return
	}

	s.Reply(msg, reply)
	s.recalculateTimer()
}

func (s *Scheduler) recalculateTimer() {
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}

	var err error
	s.nextTask, err = s.DB.GetNextTask()
	if err != nil {
		if err != sql.ErrNoRows {
			s.Log.Errorln("[scheduler] recalculate:", err)
		}
		return
	}

	dur := s.nextTask.Time.Sub(time.Now())
	s.Log.Debugln("[scheduler] next task in", dur)
	if dur < 0 {
		s.taskFinished()
		s.recalculateTimer()
		return
	}
	s.timer = time.AfterFunc(dur, s.taskFinished)
}

func (s *Scheduler) taskFinished() {
	t := s.nextTask
	t.FinishedOn = time.Now()
	s.Log.Infoln("[scheduler] task finished:", t)

	if err := s.DB.StoreTask(t); err != nil {
		s.Log.Errorln("[scheduler] could not store finished task: ", err)
	}
	s.Publish(s.nextTask.Reply)
	s.recalculateTimer()
}
