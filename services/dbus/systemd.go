// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package dbus

import "github.com/godbus/dbus"

type LogindObject struct {
	dbus.BusObject
}

func NewLogindObject(conn *dbus.Conn) *LogindObject {
	return &LogindObject{
		conn.Object("org.freedesktop.login1", "/org/freedesktop/login1"),
	}
}

func (o *LogindObject) PowerOff() error {
	return o.Call("org.freedesktop.login1.Manager.PowerOff", 0, false).Err
}
