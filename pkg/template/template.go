// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package template implements data-driven templates for generating stark messages.
//
// It follows the syntax of the standard library package text/templates.
package template

import (
	"bytes"
	"encoding/json"
	"html/template"

	"github.com/xconstruct/stark/proto"
)

type Template struct {
	tpl *template.Template
}

func New() *Template {
	return &Template{
		template.New(""),
	}
}

func (t *Template) Parse(text string) (*Template, error) {
	_, err := t.tpl.Parse("<script>" + text + "</script>")
	return t, err
}

func (t *Template) Execute(msg *proto.Message, data interface{}) error {
	var b bytes.Buffer
	if err := t.tpl.Execute(&b, data); err != nil {
		return err
	}
	by := b.Bytes()
	by = by[8 : len(by)-9] // Strip <script> tags.
	return json.Unmarshal(by, &msg)
}

func MessageToData(msg proto.Message) interface{} {
	data := make(map[string]interface{})
	msg.DecodePayload(&data)
	return data
}
