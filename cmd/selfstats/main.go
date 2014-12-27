// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/xconstruct/stark/core"
	"github.com/xconstruct/stark/pkg/selfspy"
)

var activities = []selfspy.ActivityDef{
	{"iceweasel", "reddit", "browse/reddit", 1},
	{"iceweasel", "news.ycombinator.com", "browse/hn", 1},
	{"iceweasel", "new tab", "browse/nothing", -1},
	{"terminal", "github.com/xconstruct/stark", "programming/stark", 3},
	{"terminal", "~/code", "programming", 3},
	{"vlc", "", "watch", 5},
}

func main() {
	app := core.NewApp("stark", "client")
	db, err := core.OpenDatabase(core.DatabaseConfig{
		Driver: "sqlite3",
		Source: os.Getenv("HOME") + "/.selfspy/selfspy.sqlite",
	})
	app.Must(err)

	interval := 15 * time.Minute
	before := time.Now().Truncate(interval)

	for {
		var keys []selfspy.Keys
		var clicks []selfspy.Click
		after := before.Add(-interval)
		if err := db.Where("created_at BETWEEN ? AND ?", after, before).Find(&keys).Error; err != nil {
			if err != gorm.RecordNotFound {
				app.Log.Fatal(err)
			}
		}
		if err := db.Where("created_at BETWEEN ? AND ?", after, before).Find(&clicks).Error; err != nil {
			if err != gorm.RecordNotFound {
				app.Log.Fatal(err)
			}
		}

		acts := selfspy.Summarize(activities, selfspy.ToEvents(keys, clicks))
		if len(acts) > 0 {
			fmt.Printf("%s: %v\n", after, acts[0])
		}

		before = after
	}
}
