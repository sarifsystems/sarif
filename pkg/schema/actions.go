// Copyright (C) 2014 Constantin Schomburg <me@cschomburg.com>
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package schema

type Action struct {
	*Thing
	Reply         string      `json:"reply,omitempty"`
	ReplyNegative string      `json:"reply_neg,omitempty"`
	Name          string      `json:"name,omitempty"`
	Payload       interface{} `json:"p,omitempty"`
}

type MultipleChoiceAction struct {
	Action
	Choices map[string]interface{} `json:"choices,omitempty"`
}

type TextEntryAction Action
type ConfirmAction Action
type DeleteAction Action
type CancelAction Action
