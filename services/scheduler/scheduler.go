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
	"github.com/xconstruct/stark/util"
)

var Module = core.Module{
	Name:        "scheduler",
	Version:     "1.0",
	NewInstance: newInstance,
}

func init() {
	core.RegisterModule(Module)
}

func newInstance(ctx *core.Context) (core.ModuleInstance, error) {
	return NewService(ctx)
}

type Scheduler struct {
	timer    *time.Timer
	DB       Database
	ctx      *core.Context
	proto    *proto.Client
	nextTask Task
}

func NewService(ctx *core.Context) (*Scheduler, error) {
	s := &Scheduler{
		DB:    &sqlDatabase{ctx.Database.Driver(), ctx.Database.DB},
		ctx:   ctx,
		proto: proto.NewClient("scheduler", ctx.Proto),
	}
	return s, nil
}

func (s *Scheduler) Enable() error {
	if err := s.DB.Setup(); err != nil {
		return err
	}
	if err := s.proto.Subscribe("schedule", "", s.handle); err != nil {
		return err
	}
	s.recalculateTimer()
	return nil
}
func (s *Scheduler) Disable() error { return nil }

func (s *Scheduler) handle(msg proto.Message) {
	if !msg.IsAction("schedule") {
		return
	}

	var t Task
	if err := msg.DecodePayload(&t); err != nil {
		s.ctx.Log.Warnln("[scheduler] received bad payload:", err)
		s.publish(msg.Reply(proto.BadRequest(err)))
		return
	}

	if t.Time.IsZero() {
		t.Time = time.Now()
	}
	if t.Duration != "" {
		dur, err := util.ParseDuration(t.Duration)
		if err != nil {
			s.ctx.Log.Warnln("[scheduler] received bad payload:", err)
			s.publish(msg.Reply(proto.BadRequest(err)))
			return
		}
		t.Time = t.Time.Add(dur)
	}
	if t.Reply.Action == "" {
		text := msg.Text
		if text == "" {
			text = "Reminder from " + util.FuzzyTime(time.Now()) + " finished."
		}
		t.Reply = proto.Message{
			Action:      "schedule/finished",
			Destination: msg.Source,
			Text:        text,
		}
	}
	if t.Reply.CorrId == "" {
		t.Reply.CorrId = msg.Id
	}
	s.ctx.Log.Infoln("[scheduler] new task:", t)

	if err := s.DB.StoreTask(t); err != nil {
		s.ctx.Log.Errorln("[scheduler] could not store task:", err)
		s.publish(msg.Reply(proto.InternalError(err)))
		return
	}

	reply := proto.Message{Action: "schedule/created"}
	if err := reply.EncodePayload(t); err != nil {
		s.ctx.Log.Errorln("[scheduler] could not encode reply:", err)
		s.publish(msg.Reply(proto.InternalError(err)))
		return
	}

	s.publish(msg.Reply(reply))
	s.recalculateTimer()
}

func (s *Scheduler) publish(msg proto.Message) {
	if err := s.proto.Publish(msg); err != nil {
		s.ctx.Log.Errorln(err)
	}
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
			s.ctx.Log.Errorln("[scheduler] recalculate:", err)
		}
		return
	}

	dur := s.nextTask.Time.Sub(time.Now())
	s.ctx.Log.Debugln("[scheduler] next task in", dur)
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
	s.ctx.Log.Infoln("[scheduler] task finished:", t)

	if err := s.DB.StoreTask(t); err != nil {
		s.ctx.Log.Errorln("[scheduler] could not store finished task: ", err)
	}
	s.publish(s.nextTask.Reply)
	s.recalculateTimer()
}
