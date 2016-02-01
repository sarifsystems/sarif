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

	"github.com/jinzhu/gorm"
	"github.com/xconstruct/stark/pkg/util"
	"github.com/xconstruct/stark/proto"
	"github.com/xconstruct/stark/services"
)

var Module = &services.Module{
	Name:        "scheduler",
	Version:     "1.0",
	NewInstance: NewService,
}

type Dependencies struct {
	DB     *gorm.DB
	Log    proto.Logger
	Client *proto.Client
}

type Scheduler struct {
	DB  *gorm.DB
	Log proto.Logger
	*proto.Client

	mutex    sync.Mutex
	timer    *time.Timer
	nextTask Task
}

func NewService(deps *Dependencies) *Scheduler {
	return &Scheduler{
		DB:     deps.DB,
		Log:    deps.Log,
		Client: deps.Client,
	}
}

func (s *Scheduler) Enable() error {
	rand.Seed(time.Now().UnixNano())
	createIndizes := !s.DB.HasTable(&Task{})
	if err := s.DB.AutoMigrate(&Task{}).Error; err != nil {
		return err
	}
	if createIndizes {
		if err := s.DB.Model(&Task{}).AddIndex("time", "time").Error; err != nil {
			return err
		}
	}
	if err := s.Subscribe("schedule", "", s.handle); err != nil {
		return err
	}
	go s.simpleCron()
	go s.recalculateTimer()
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
			s.Log.Warnln("[scheduler] received bad payload:", err)
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

	if err := s.DB.Save(&t.Task).Error; err != nil {
		s.Log.Errorln("[scheduler] could not store task:", err)
		s.ReplyInternalError(msg, err)
		return
	}

	reply := proto.Message{Action: "schedule/created"}
	if err := reply.EncodePayload(t.Task); err != nil {
		s.Log.Errorln("[scheduler] could not encode reply:", err)
		s.ReplyInternalError(msg, err)
		return
	}

	s.Reply(msg, reply)
	go s.recalculateTimer()
}

func (s *Scheduler) GetNextTask() (t Task, err error) {
	err = s.DB.Where("finished = ?", false).Order("time ASC").Limit(1).First(&t).Error
	return
}

func (s *Scheduler) recalculateTimer() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}

	var err error
	s.nextTask, err = s.GetNextTask()
	if err != nil {
		if err != gorm.RecordNotFound {
			s.Log.Errorln("[scheduler] recalculate:", err)
		}
		return
	}

	dur := s.nextTask.Time.Sub(time.Now())
	s.Log.Debugln("[scheduler] next task in", dur)
	if dur < 0 {
		s.taskFinished()
		return
	}
	s.timer = time.AfterFunc(dur, s.taskFinished)
}

func (s *Scheduler) taskFinished() {
	t := s.nextTask
	t.Finished = true
	s.Log.Infoln("[scheduler] task finished:", t)

	if err := s.DB.Save(&t).Error; err != nil {
		s.Log.Errorln("[scheduler] could not store finished task: ", err)
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
		s.Publish(proto.CreateMessage(action, nil))
	}
}
