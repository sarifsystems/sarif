// Copyright (C) 2016 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package dbus

import (
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/godbus/dbus"
)

type MprisPlayer struct {
	Id       string `json:"-"`
	Name     string `json:"name"`
	Identity string `json:"identity"`
	Status   string `json:"status"`

	Title  string `json:"title,omitempty"`
	Artist string `json:"artist,omitempty"`
	Album  string `json:"album,omitempty"`
	Url    string `json:"url,omitempty"`
}

func (p *MprisPlayer) UpdateProperties(props map[string]dbus.Variant) {
	changed := false
	if v, ok := props["PlaybackStatus"]; ok {
		p.Status = strings.ToLower(v.Value().(string))
		changed = true
	}
	if v, ok := props["Metadata"]; ok {
		m := v.Value().(map[string]dbus.Variant)
		if vv, ok := m["xesam:title"]; ok {
			p.Title = vv.Value().(string)
			changed = true
		}
		if vv, ok := m["xesam:url"]; ok {
			p.Url = vv.Value().(string)
			changed = true
		}
		if vv, ok := m["xesam:artist"]; ok {
			p.Artist = vv.Value().([]string)[0]
			changed = true
		}
		if vv, ok := m["xesam:album"]; ok {
			p.Album = vv.Value().(string)
			changed = true
		}
	}

	if changed {
		spew.Dump(p)
	}
}
