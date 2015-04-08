// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package web

import (
	"strings"
	"testing"
)

func TestExtract(t *testing.T) {
	thing := map[string]interface{}{
		"somekey": map[string]interface{}{
			"anotherkey": 3,
			"a_string":   "this is string",
		},
		"second": []interface{}{
			"whelp",
			5,
			map[string]interface{}{
				"object": "shiny",
			},
		},
	}

	test := func(path string, expected interface{}) {
		p := strings.Split(path, "/")
		x, err := extract(thing, p)
		if err != nil {
			t.Error(err)
		} else if x != expected {
			t.Errorf("key '%s': expected %v, got %v", path, expected, x)
		}
	}

	test("somekey/anotherkey", 3)
	test("somekey/a_string", "this is string")
	test("second/first", "whelp")
	test("second/1", 5)
	test("second/last/object", "shiny")
}

func TestParseActionAsURL(t *testing.T) {
	tests := map[string]jsonRequest{
		"json/get/http/host.com/path/subpath?query=things&test=three#facts/moo": {
			Method:  "get",
			Url:     "http://host.com/path/subpath?query=things&test=three",
			Extract: "facts/moo",
		},
		"json/get/http/catfacts-api.appspot.com/api/facts#facts/first": {
			Method:  "get",
			Url:     "http://catfacts-api.appspot.com/api/facts",
			Extract: "facts/first",
		},
		"json/api.github.com/repos/xconstruct/stark/commits#first/commit/message": {
			Method:  "get",
			Url:     "https://api.github.com/repos/xconstruct/stark/commits",
			Extract: "first/commit/message",
		},
	}

	for action, exp := range tests {
		req, err := parseActionAsURL(action)
		if err != nil {
			t.Fatal(err)
		}
		if req.Method != exp.Method {
			t.Errorf("%s: expected '%s', got '%s'", action, exp.Method, req.Method)
		}
		if req.Url != exp.Url {
			t.Errorf("%s: expected '%s', got '%s'", action, exp.Url, req.Url)
		}
		if req.Extract != exp.Extract {
			t.Errorf("%s: expected '%s', got '%s'", action, exp.Extract, req.Extract)
		}
	}
}
