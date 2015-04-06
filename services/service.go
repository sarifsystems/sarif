// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package services

type Module struct {
	Name        string
	Version     string
	Description string

	NewInstance interface{}
}

type Config interface {
	Exists() bool
	Set(v interface{}) error
	Get(v interface{}) (error, bool)

	Dir() string
}
