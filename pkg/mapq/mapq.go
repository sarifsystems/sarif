// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package mapq provides aggregation and filtering functions for generic maps
package mapq

import (
	"fmt"
	"math"
	"sort"
)

type M map[string]interface{}
type Collection []map[string]interface{}
type Group map[string][]map[string]interface{}

type Filter map[string]interface{}

func (m M) Matches(filter Filter) bool {
	if filter == nil {
		return true
	}
	for k, v := range filter {
		key, op := splitQueryOp(k)

		if !Matches(m[key], op, v) {
			return false
		}
	}
	return true
}

func (m M) stringKey(key string) string {
	return fmt.Sprintf("%v", m[key])
}

func (m M) floatKey(key string) (float64, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	return toFloat(v)
}

func (m M) Map() map[string]interface{} {
	return map[string]interface{}(m)
}

func ToCollection(v interface{}) Collection {
	var ok bool

	switch sl := v.(type) {
	case []map[string]interface{}:
		return Collection(sl)
	case []interface{}:
		c := make([]map[string]interface{}, len(sl))
		for i, m := range sl {
			if c[i], ok = m.(map[string]interface{}); !ok {
				return nil
			}
		}
		return Collection(c)
	}
	return nil
}

func (c Collection) Any(filter Filter) bool {
	for _, m := range c {
		if m != nil && M(m).Matches(filter) {
			return true
		}
	}
	return false
}

func (c Collection) All(filter Filter) bool {
	for _, m := range c {
		if m == nil || !M(m).Matches(filter) {
			return false
		}
	}
	return true
}

func (c Collection) Filter(filter Filter) Collection {
	c2 := Collection{}
	for _, m := range c {
		if m != nil && M(m).Matches(filter) {
			c2 = append(c2, m)
		}
	}
	return c2
}

func (c Collection) Slice() []map[string]interface{} {
	return []map[string]interface{}(c)
}

func (c Collection) GroupBy(key string) Group {
	g := Group{}
	for _, m := range c {
		kv := M(m).stringKey(key)
		g[kv] = append(g[kv], m)
	}
	return g
}

type sortableCollection struct {
	C   Collection
	Key string
	Op  string
}

func (s *sortableCollection) Len() int      { return len(s.C) }
func (s *sortableCollection) Swap(i, j int) { s.C[i], s.C[j] = s.C[j], s.C[i] }
func (s *sortableCollection) Less(i, j int) bool {
	a, b := M(s.C[i]), M(s.C[j])
	return Matches(a, s.Op, b)
}

func (c Collection) OrderBy(key string, order string) Collection {
	op := "<"
	if order == "desc" || order == "DESC" {
		op = ">"
	}
	s := &sortableCollection{c, key, op}
	sort.Sort(s)
	return c
}

func (c Collection) Aggregate(key string, op string) float64 {
	var first bool
	var f float64
	for _, m := range c {
		v, ok := M(m).floatKey(key)
		if !ok {
			continue
		}
		switch op {
		case "count":
			f += 1
		case "sum":
			f += v
		case "mean":
			f += v
		case "min":
			if first {
				f, first = v, false
			}
			f = math.Min(f, v)
		case "max":
			if first {
				f, first = v, false
			}
			f = math.Max(f, v)
		}
	}
	if op == "mean" {
		f /= float64(len(c))
	}
	return f
}

func (c Collection) Each(f func(m M) M) Collection {
	for k, m := range c {
		c[k] = f(M(m)).Map()
	}
	return c
}

func (g Group) Each(f func(c Collection) Collection) Group {
	for k, c := range g {
		g[k] = f(c)
	}
	return g
}

func (g Group) Aggregate(key, op string) M {
	m := make(map[string]interface{})
	for k, c := range g {
		m[k] = Collection(c).Aggregate(key, op)
	}
	return m
}

func (g Group) Map() map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range g {
		m[k] = v
	}
	return m
}

func (g Group) M() M {
	return M(g.Map())
}
