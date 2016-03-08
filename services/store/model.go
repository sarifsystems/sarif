// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package store

import (
	"crypto/rand"
	"errors"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	ErrNoResult = errors.New("No result found.")
)

type Document struct {
	Id    int64     `json:"-"`
	Key   string    `json:"key" sql:"index" gorm:"column:dkey"`
	Value []byte    `json:"value,omitempty"`
	Time  time.Time `json:"time" sql:"index"`
}

func (Document) TableName() string {
	return "store"
}

func (doc Document) String() string {
	return "Document " + doc.Key + "."
}

type Store interface {
	Setup() error
	Put(doc Document) (Document, error)
	Get(key string) (Document, error)
	Del(key string) error
	Scan(prefix, start, end string) ([]string, error)
}

type sqlStore struct {
	DB *gorm.DB
}

func (d *sqlStore) Setup() error {
	return d.DB.AutoMigrate(&Document{}).Error
}

func generateId() string {
	const alphanum = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	var bytes = make([]byte, 12)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func (d *sqlStore) Put(doc Document) (Document, error) {
	if doc.Key == "" {
		doc.Key = generateId()
	} else if strings.HasSuffix(doc.Key, "$") {
		doc.Key = strings.TrimSuffix(doc.Key, "$")
		doc.Key += generateId()
	} else {
		if err := d.Del(doc.Key); err != nil {
			return doc, err
		}
	}
	doc.Time = time.Now()

	err := d.DB.Save(&doc).Error
	return doc, err
}

func (d *sqlStore) Del(key string) error {
	return d.DB.Where(&Document{Key: key}).Delete(&Document{}).Error
}

func (d *sqlStore) Get(key string) (Document, error) {
	var doc Document
	doc.Key = key
	err := d.DB.Where(&doc).Find(&doc).Error
	if err == gorm.ErrRecordNotFound {
		err = ErrNoResult
	}
	return doc, err
}

func (d *sqlStore) Scan(prefix, start, end string) ([]string, error) {
	var keys []string
	db := d.DB.Model(&Document{})
	if prefix != "" {
		db = db.Where("dkey LIKE ?", prefix+"%")
	}
	if start != "" {
		db = db.Where("dkey > ?", start)
	}
	if end != "" {
		db = db.Where("dkey < ?", end)
	}

	err := db.Pluck("dkey", &keys).Error
	return keys, err
}
