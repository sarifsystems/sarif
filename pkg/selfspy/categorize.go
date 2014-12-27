// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package selfspy

import (
	"sort"
	"strings"
)

type Activity struct {
	Type    string
	Process string
	Title   string
	Weight  float32
}

type ByWeight []Activity

func (a ByWeight) Len() int           { return len(a) }
func (a ByWeight) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByWeight) Less(i, j int) bool { return a[i].Weight > a[j].Weight }

func KeysToEvents(ks []Keys) []Event {
	e := make([]Event, len(ks))
	for i, k := range ks {
		e[i] = k.Event
	}
	return e
}

type ActivityDef struct {
	Process string
	Window  string
	Type    string
	Weight  float32
}

func (d ActivityDef) Matches(e Event) bool {
	if !strings.Contains(strings.ToLower(e.Window.Process.Name), d.Process) {
		return false
	}
	if !strings.Contains(strings.ToLower(e.Window.Title), d.Window) {
		return false
	}
	return true
}

func Categorize(defs []ActivityDef, e Event) Activity {
	a := Activity{
		Process: e.Window.Process.Name,
		Title:   e.Window.Title,
		Weight:  0.1,
	}

	for _, def := range defs {
		if def.Matches(e) {
			a.Type = def.Type
			a.Weight = def.Weight
			return a
		}
	}
	a.Type = "unknown/" + a.Process
	return a
}

func ToEvents(keys []Keys, clicks []Click) []Event {
	e := make([]Event, 0)
	for _, k := range keys {
		e = append(e, k.Event)
	}
	for _, k := range clicks {
		e = append(e, k.Event)
	}
	return e
}

func Summarize(defs []ActivityDef, es []Event) []Activity {
	acts := make(map[string]*Activity)
	for _, e := range es {
		a := Categorize(defs, e)
		if b, ok := acts[a.Type]; ok {
			b.Weight += a.Weight
		} else {
			acts[a.Type] = &a
		}
	}

	as := make([]Activity, 0, len(acts))
	for _, a := range acts {
		as = append(as, *a)
	}
	sort.Sort(ByWeight(as))
	return as
}
