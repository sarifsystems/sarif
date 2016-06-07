// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mapq

import "strings"

var ops = []string{
	"==",
	"!=",
	">",
	"<",
	">=",
	"<=",
	"^",
	"$",
}

func splitQueryOp(key string) (q, op string) {
	for _, op := range ops {
		if strings.HasSuffix(key, " "+op) {
			return strings.TrimSuffix(key, " "+op), op
		}
	}
	return key, ""
}

func Matches(a interface{}, op string, b interface{}) bool {
	if op == "" {
		op = "=="
	}

	if bm, ok := b.(map[string]interface{}); ok {
		b = Filter(bm)
	}
	if am, ok := a.(map[string]interface{}); ok {
		a = M(am)
	}
	if bm, ok := b.(Filter); ok {
		if am, ok := a.(M); ok {
			if op == "==" {
				return am.Matches(bm)
			} else if op == "!=" {
				return !am.Matches(bm)
			}
		}
		return false
	}

	if op == "^" {
		if av, ok := a.(string); ok {
			if bv, ok := b.(string); ok {
				return strings.HasPrefix(av, bv)
			}
			return false
		}
	}
	if op == "$" {
		if av, ok := a.(string); ok {
			if bv, ok := b.(string); ok {
				return strings.HasSuffix(av, bv)
			}
			return false
		}
	}

	c, ok := compare(a, b)
	if !ok {
		return false
	}
	switch op {
	case "":
		fallthrough
	case "==":
		return c == 0
	case "!=":
		return c != 0
	case ">":
		return c > 0
	case "<":
		return c < 0
	case ">=":
		return c >= 0
	case "<=":
		return c >= 0
	}

	return false
}

func compare(a, b interface{}) (int, bool) {
	if a == b {
		return 0, true
	}

	if av, ok := a.(string); ok {
		if bv, ok := b.(string); ok {
			return strings.Compare(av, bv), true
		}
		return 0, false
	}
	if av, ok := toInt(a); ok {
		if bv, ok := toInt(b); ok {
			c := av - bv
			if c > 0 {
				return 1, true
			} else if c < 0 {
				return -1, true
			} else {
				return 0, true
			}
		}
	}

	if av, ok := toFloat(a); ok {
		if bv, ok := toFloat(b); ok {
			c := av - bv
			if c > 0 {
				return 1, true
			} else if c < 0 {
				return -1, true
			} else {
				return 0, true
			}
		}
	}
	return 0, false
}

func toInt(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case uint:
		return int64(n), true
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		return int64(n), true
	}
	return 0, false
}

func toFloat(v interface{}) (float64, bool) {
	if n, ok := toInt(v); ok {
		return float64(n), true
	}

	switch n := v.(type) {
	case float32:
		return float64(n), true
	case float64:
		return float64(n), true
	}
	return 0, false
}
