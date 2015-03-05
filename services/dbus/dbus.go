// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package dbus

import "github.com/godbus/dbus"

type NotificationObject struct {
	*dbus.Object
}

func NewNotificationObject(conn *dbus.Conn) *NotificationObject {
	return &NotificationObject{
		conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications"),
	}
}

func (o *NotificationObject) Notify(summary, body string) error {
	method := "org.freedesktop.Notifications.Notify"
	name := "kipp"
	replacesId := uint32(0)
	icon := ""
	actions := []string{}
	hints := map[string]dbus.Variant{}
	timeout := int32(5000)

	c := o.Call(method, 0, name, replacesId, icon, summary, body, actions, hints, timeout)
	return c.Err
}
