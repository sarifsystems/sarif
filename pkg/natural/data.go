// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package natural

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

var reSplitData = regexp.MustCompile(`([^\[]+)(\[.+\])`)
var reDataMatchVars = regexp.MustCompile(`\[([\w=:, ]+)\]`)
var reVar = regexp.MustCompile(`(\w+)(:\w+)?=(.+)`)

type DataSet []Data

type Data struct {
	Action   string
	Sentence string
	Vars     []*Var
}

func (d Data) CleanedSentence(placeholder string) string {
	s := d.Sentence
	for _, v := range d.Vars {
		p := placeholder
		switch p {
		case "[name]":
			p = "[" + v.Name + "]"
		case "[var]":
			p = v.String()
		}
		s = strings.Replace(s, v.Value, p, -1)
	}
	return s
}

type Var struct {
	Name   string  `json:"name"`
	Type   string  `json:"type,omitempty"`
	Value  string  `json:"value"`
	Weight float64 `json:"weight"`
}

func (v Var) String() string {
	s := ""
	if v.Name != "" {
		s += v.Name
	}
	if v.Type != "" {
		s += ":" + v.Type
	}
	if v.Value != "" {
		s += "=" + v.Value
	}
	return "[" + s + "]"
}

type ParseError struct {
	Line int
	Err  error
}

func (e ParseError) Error() string {
	return fmt.Sprintf("Parse error on line %d: %s", e.Line, e.Err.Error())
}

func ReadDataSet(r io.Reader) (DataSet, error) {
	var set DataSet
	s := bufio.NewScanner(r)

	i := 0
	action := ""
	for s.Scan() {
		i++
		line := s.Text()

		if line == "" {
			action = ""
			continue
		}
		if action == "" {
			action = line
			continue
		}

		data, err := ReadData(line)
		if err != nil {
			return set, &ParseError{i, err}
		}
		data.Action = action
		set = append(set, data)
	}

	return set, nil
}

func ReadData(data string) (Data, error) {
	d := Data{
		Vars: make([]*Var, 0),
	}

	m := reSplitData.FindStringSubmatch(data)
	if m == nil {
		d.Sentence = data
	} else {
		d.Sentence = m[1]

		groups := reDataMatchVars.FindAllStringSubmatch(m[2], -1)
		for _, group := range groups {
			vars := strings.Split(group[1], ",")
			for _, vs := range vars {
				v := reVar.FindStringSubmatch(vs)
				if v == nil {
					return d, fmt.Errorf("unexpected variable format %q", vs)
				}
				d.Vars = append(d.Vars, &Var{
					Name:  v[1],
					Type:  strings.TrimLeft(v[2], ":"),
					Value: v[3],
				})
			}
		}
	}

	d.Sentence = strings.TrimSpace(d.Sentence)
	return d, nil
}
