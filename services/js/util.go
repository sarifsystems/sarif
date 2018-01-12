// Copyright (C) 2018 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package js

import (
	"encoding/json"

	"github.com/robertkrimen/otto"
	"github.com/sarifsystems/sarif/sarif"
)

func objectGetString(o *otto.Object, key string) string {
	v, err := o.Get(key)
	if err != nil || !v.IsDefined() {
		return ""
	}
	s, _ := v.ToString()
	return s
}

func objectToMessage(v otto.Value) sarif.Message {
	msg := sarif.Message{}

	o := v.Object()
	if o == nil {
		return msg
	}

	msg.Version = objectGetString(o, "sarif")
	msg.Id = objectGetString(o, "id")
	msg.Action = objectGetString(o, "action")
	msg.Source = objectGetString(o, "src")
	msg.Destination = objectGetString(o, "dst")
	msg.CorrId = objectGetString(o, "corr")
	msg.Text = objectGetString(o, "text")

	p, err := o.Get("p")
	if err == nil && p.IsDefined() {
		v, _ := p.Export()
		msg.EncodePayload(v)
	}

	return msg
}

func messageToObject(otto *otto.Otto, msg sarif.Message) otto.Value {
	js, _ := json.Marshal(msg)
	v, _ := otto.Run("(" + string(js) + ")")
	return v
}
