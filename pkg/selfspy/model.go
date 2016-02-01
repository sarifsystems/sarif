// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package selfspy provides access to the selfspy database.
package selfspy

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Window struct {
	Id        int64
	CreatedAt time.Time
	Title     string
	ProcessId int64

	Process Process
}

func (w *Window) AfterFind(tx *gorm.DB) error {
	return tx.Model(w).Related(&w.Process).Error
}

func (Window) TableName() string {
	return "window"
}

type Process struct {
	Id        int64
	CreatedAt time.Time
	Name      string
}

func (Process) TableName() string {
	return "process"
}

type Event struct {
	Id        int64
	CreatedAt time.Time
	WindowId  int64

	Window Window
}

type Click struct {
	Event
	Button  int
	Press   bool
	X       int
	Y       int
	Nrmoves int
}

func (c *Click) AfterFind(tx *gorm.DB) error {
	return tx.Model(c).Related(&c.Window).Error
}

func (Click) TableName() string {
	return "click"
}

type Keys struct {
	Event
	//Text       []byte
	Started time.Time
	Nrkeys  int
	//Keys       []byte
	//Timings    []byte
}

func (k *Keys) AfterFind(tx *gorm.DB) error {
	return tx.Model(k).Related(&k.Window).Error
}

func (Keys) TableName() string {
	return "keys"
}
