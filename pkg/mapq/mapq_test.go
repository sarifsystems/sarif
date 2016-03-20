// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package mapq

import "testing"

func TestMatches(t *testing.T) {
	example := map[string]interface{}{
		"name": "John Smith",
		"age":  43,
		"address": map[string]interface{}{
			"street": "Broadway St.",
			"number": 23,
		},
		"relevance": 1337.5443,
	}

	ok := M(example).Matches(Filter{
		"name ^": "John",
	})
	if !ok {
		t.Error("name should match")
	}

	ok = M(example).Matches(Filter{
		"name": "John Smith",
	})
	if !ok {
		t.Error("name should match")
	}

	ok = M(example).Matches(Filter{
		"name": "Brad",
	})
	if ok {
		t.Error("name should not match")
	}

	ok = M(example).Matches(Filter{
		"age >": 42,
	})
	if !ok {
		t.Error("age should match")
	}

	ok = M(example).Matches(Filter{
		"age !=": "5",
	})
	if ok {
		t.Error("age should not match")
	}

	ok = M(example).Matches(Filter{
		"name >": "Adam",
		"name <": "Kaylee",
		"name ^": "John",
		"name $": "Smith",

		"age":   43,
		"age <": 43.5,
		"age >": -1,

		"relevance >": 1337,
	})
	if !ok {
		t.Error("big filter should match")
	}

	ok = M(example).Matches(Filter{
		"address": Filter{
			"street ^": "Broadway",
		},
	})
	if !ok {
		t.Error("deep filter should match")
	}

	ok = M(example).Matches(Filter{
		"address !=": Filter{
			"street $": "Avenue",
		},
	})
	if !ok {
		t.Error("negated deep filter should match")
	}
}
