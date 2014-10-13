// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mood

import "time"

const TagScoreUndefined = 9999

type Record struct {
	Id        int64     `json:"-"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Score     int       `json:"score"`
	Text      string    `json:"text"`
	Tags      []Tag     `json:"tags" gorm:"many2many:mood_record_tags"`
}

func (r *Record) RecalculateScore() {
	r.Score = 0
	for _, tag := range r.Tags {
		if tag.Score != TagScoreUndefined {
			r.Score += tag.Score
		}
	}
}

func (Record) TableName() string {
	return "mood_records"
}

type Tag struct {
	Id    int64  `json:"-"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

func (Tag) TableName() string {
	return "mood_tags"
}
