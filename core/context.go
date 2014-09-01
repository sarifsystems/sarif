// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/xconstruct/stark/conf"
	"github.com/xconstruct/stark/database"
	"github.com/xconstruct/stark/log"
	"github.com/xconstruct/stark/proto"
)

type Context struct {
	Database *database.DB
	Log      *log.Logger
	Proto    proto.Endpoint
	Config   *conf.Config
}

func (ctx *Context) Must(err error) {
	if err != nil {
		ctx.Log.Errorln(err)
	}
}
